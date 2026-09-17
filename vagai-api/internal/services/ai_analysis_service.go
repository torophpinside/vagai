package services

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/anomalyco/vagai-api/internal/models"
	"gorm.io/gorm"
)

type AIAnalysisService struct {
	db *gorm.DB
}

func NewAIAnalysisService(db *gorm.DB) *AIAnalysisService {
	return &AIAnalysisService{db: db}
}

// CreateAnalysisResult creates a new AI analysis record with a pending status.
func (s *AIAnalysisService) CreateAnalysisResult(ctx context.Context, orgID, prepID, questionID uint) (*models.AIAnalysisResult, error) {
	result := &models.AIAnalysisResult{
		OrganizationID: orgID,
		PreparationID:  prepID,
		QuestionID:     questionID,
		Status:         models.AIAnalysisStatusPending,
	}

	if err := s.db.WithContext(ctx).Create(result).Error; err != nil {
		return nil, fmt.Errorf("failed to create AI analysis result: %w", err)
	}

	return result, nil
}

// GetAnalysisResultsForPreparation retrieves all analysis results for a given preparation.
func (s *AIAnalysisService) GetAnalysisResultsForPreparation(ctx context.Context, prepID uint) ([]models.AIAnalysisResult, error) {
	var results []models.AIAnalysisResult
	if err := s.db.WithContext(ctx).Where("preparation_id = ?", prepID).Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get AI analysis results: %w", err)
	}
	return results, nil
}

// UpdateAnalysisResultStatus updates the status, text, score and source of an analysis result.
func (s *AIAnalysisService) UpdateAnalysisResultStatus(ctx context.Context, resultID uint, status models.AIAnalysisStatus, text string, score float64, source models.VerificationSource) error {
	err := s.db.WithContext(ctx).Model(&models.AIAnalysisResult{}).Where("id = ?", resultID).Updates(map[string]interface{}{
		"status":         status,
		"analysis_text":  text,
		"analysis_score": score,
		"source":         source,
	}).Error

	if err != nil {
		return fmt.Errorf("failed to update AI analysis result: %w", err)
	}
	return nil
}

// StartAsyncAnalysis manages the background process of analyzing all questions in a preparation.
// Existing results for the preparation are replaced: each new correction
// request substitutes the previous per-question feedback.
func (s *AIAnalysisService) StartAsyncAnalysis(ctx context.Context, prepID uint, questions []models.InterviewQuestion) {
	if len(questions) == 0 {
		return
	}
	// Use a separate context for background processing to avoid cancellation when the HTTP request ends.
	bgCtx := context.Background()

	go func() {
		log.Printf("[AIAnalysis] Starting async analysis for preparation %d", prepID)

		// Replace previous feedback for this preparation (same org scope as its questions).
		orgID := questions[0].OrganizationID
		if err := s.db.WithContext(bgCtx).
			Where("preparation_id = ? AND organization_id = ?", prepID, orgID).
			Delete(&models.AIAnalysisResult{}).Error; err != nil {
			log.Printf("[AIAnalysis] Error clearing previous results for preparation %d: %v", prepID, err)
		}

		for _, q := range questions {
			// We need to fetch the answer from PracticeEntry since it's not in InterviewQuestion.
			// Only answered questions are analyzed; unanswered ones get no feedback block.
			var entry models.PracticeEntry
			if err := s.db.Where("question_id = ? AND organization_id = ?", q.ID, q.OrganizationID).First(&entry).Error; err != nil {
				log.Printf("[AIAnalysis] Answer not found for question %d: %v", q.ID, err)
				continue
			}
			if strings.TrimSpace(entry.Answer) == "" {
				continue
			}

			// 1. Create pending result
			res, err := s.CreateAnalysisResult(bgCtx, q.OrganizationID, prepID, q.ID)
			if err != nil {
				log.Printf("[AIAnalysis] Error creating result for question %d: %v", q.ID, err)
				continue
			}

			// 2. Call AI service with individual retries per question.
			score, feedback, source, err := analyzeQuestionWithRetry(q, entry)
			if err != nil {
				log.Printf("[AIAnalysis] Analysis failed for question %d: %v", q.ID, err)
				s.UpdateAnalysisResultStatus(bgCtx, res.ID, models.AIAnalysisStatusError, fmt.Sprintf("AI Error: %v", err), 0, "")
				continue
			}

			// 3. Update result as completed
			if err := s.UpdateAnalysisResultStatus(bgCtx, res.ID, models.AIAnalysisStatusCompleted, feedback, score, source); err != nil {
				log.Printf("[AIAnalysis] Error updating result for question %d: %v", q.ID, err)
			}
		}
		log.Printf("[AIAnalysis] Completed async analysis for preparation %d", prepID)
	}()

}

// analyzeQuestionAttempts é o nº de tentativas de IA por pergunta antes do fallback.
const analyzeQuestionAttempts = 3

// analyzeQuestionWithRetry corrige uma resposta com a IA (nota real 0-10),
// com re-tentativas individuais por pergunta. Esgotadas as tentativas, usa o
// fallback heurístico determinístico (cobertura do guia, source=template) para
// nunca deixar a pergunta sem feedback. Último recurso: erro.
func analyzeQuestionWithRetry(q models.InterviewQuestion, entry models.PracticeEntry) (float64, string, models.VerificationSource, error) {
	var err error
	for attempt := 1; attempt <= analyzeQuestionAttempts; attempt++ {
		var score float64
		var feedback string
		score, feedback, err = AnalyzeQuestionAnswerScored(q.Text, entry.Answer)
		if err == nil {
			return score, feedback, models.VerificationSourceAI, nil
		}
		log.Printf("[AIAnalysis] Attempt %d/%d failed for question %d: %v", attempt, analyzeQuestionAttempts, q.ID, err)
	}

	fb, ferr := verifyPreparationFallback([]verificationQA{{
		Question:   q,
		Answer:     entry.Answer,
		Guide:      q.AnswerGuide,
		Assessment: entry.SelfAssessment,
	}})
	if ferr != nil {
		return 0, "", "", fmt.Errorf("IA e fallback falharam: %v", err)
	}
	return fb.Score, fb.Feedback, models.VerificationSourceTemplate, nil
}
