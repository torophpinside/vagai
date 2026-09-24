package services

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/anomalyco/vagai-api/internal/models"
	"gorm.io/gorm"
)

// ErrSpontaneousDescriptionRequired indica que a descrição da vaga (após trim)
// está vazia — mapeado para HTTP 400 no handler.
var ErrSpontaneousDescriptionRequired = errors.New("conteúdo da vaga é obrigatório")

const (
	// spontaneousPrepAIBudget é o orçamento de IA por requisição avulsa, menor
	// que o limite de 240s do servidor e do caminho persistido.
	spontaneousPrepAIBudget = 180 * time.Second
	// spontaneousMaxPromptRunes limita, de forma rune-safe, o trecho da
	// descrição enviado ao prompt; o texto completo permanece na resposta.
	spontaneousMaxPromptRunes = 50000
)

// SpontaneousPrepResult é o resultado em memória da preparação avulsa: nada é
// persistido. O handler o mapeia para o DTO existente (interviewPrepDTO).
type SpontaneousPrepResult struct {
	Title       string
	Company     string
	Description string
	TechStack   []string
	Questions   []models.InterviewQuestion
	Source      models.InterviewPreparationSource
	GeneratedAt time.Time
}

// GenerateSpontaneousPreparation gera, sem persistir, TargetQuestionCount
// perguntas a partir do conteúdo bruto de uma vaga. Reutiliza detecção de
// tecnologias, geração por IA (com fallback determinístico), contagem fixa e
// embaralhamento reproduzível. Stateless: não recebe DB e não cria linhas.
func GenerateSpontaneousPreparation(ctx context.Context, title, company, description, seed string) (*SpontaneousPrepResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	title = strings.TrimSpace(title)
	company = strings.TrimSpace(company)
	description = strings.TrimSpace(description)
	if description == "" {
		return nil, ErrSpontaneousDescriptionRequired
	}

	techs := DetectTechnologies(description)

	source := models.SourceTemplate
	questions := templateQuestionsToModels(BuildTemplateQuestions(techs))

	// Constituição IV: IA é aprimoramento, não dependência. Timeout, erro ou
	// saída inválida caem no template determinístico.
	aiCtx, cancel := context.WithTimeout(ctx, spontaneousPrepAIBudget)
	defer cancel()

	promptDescription := truncateRunes(description, spontaneousMaxPromptRunes)
	if aiQs, aiErr := generateInterviewQuestionsAI(aiCtx, title, company, promptDescription, techs); aiErr == nil {
		source = models.SourceAI
		questions = aiQuestionsToModels(aiQs)
	} else {
		// Apenas metadados: nunca logamos o texto da vaga.
		log.Printf("SpontaneousPrep: IA indisponível/inválida, usando template: %v", aiErr)
	}

	before := len(questions)
	questions = EnforceQuestionCount(questions, techs, seed)
	log.Printf("SpontaneousPrep (source=%s): %d -> %d perguntas (seed=%q, techs=%d)",
		source, before, len(questions), seed, len(techs))

	return &SpontaneousPrepResult{
		Title:       title,
		Company:     company,
		Description: description,
		TechStack:   techs,
		Questions:   questions,
		Source:      source,
		GeneratedAt: time.Now().UTC(),
	}, nil
}

// truncateRunes retorna um prefixo de s com no máximo max runes, sem cortar um
// caractere multibyte ao meio.
func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(s) <= max {
		return s
	}
	count := 0
	for i := range s {
		if count == max {
			return s[:i]
		}
		count++
	}
	return s
}

// CreatePersistedSpontaneousPreparation gera e PERSISTE uma preparação avulsa
// (InterviewPreparation + TargetQuestionCount InterviewQuestion) no escopo da
// organização autenticada. Reutiliza a geração pura (IA com fallback
// determinístico) e grava tudo em uma única transação: ou a preparação e todas
// as perguntas existem, ou nenhuma linha é criada. Cada chamada cria uma
// preparação nova e independente — nada é sobrescrito.
func CreatePersistedSpontaneousPreparation(db *gorm.DB, orgID uint, title, company, description, seed string) (*models.InterviewPreparation, error) {
	if orgID == 0 {
		return nil, errors.New("organization_id é obrigatório")
	}

	result, err := GenerateSpontaneousPreparation(context.Background(), title, company, description, seed)
	if err != nil {
		return nil, err
	}

	var prep models.InterviewPreparation
	err = db.Transaction(func(tx *gorm.DB) error {
		prep = models.InterviewPreparation{
			OrganizationID: orgID,
			Title:          result.Title,
			Company:        result.Company,
			Description:    result.Description,
			TechStack:      mustJSON(result.TechStack),
			Source:         result.Source,
			GeneratedAt:    result.GeneratedAt,
		}
		if err := tx.Create(&prep).Error; err != nil {
			return err
		}

		for i, q := range result.Questions {
			q.ID = 0
			q.OrganizationID = orgID
			q.PreparationID = prep.ID
			q.Position = i + 1
			if err := tx.Create(&q).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("SpontaneousPrep persist (org=%d) error: %v", orgID, err)
		return nil, err
	}

	return &prep, nil
}
