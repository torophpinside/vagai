package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/anomalyco/vagai-api/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// slugPattern valida o identificador do plano (V4).
var slugPattern = regexp.MustCompile(`^[a-z0-9-]+$`)

func parseUintParam(raw string) (uint, error) {
	v, err := strconv.ParseUint(raw, 10, 32)
	return uint(v), err
}

// planRequest é o payload de criação/edição de plano. Ponteiros distinguem
// "campo ausente" (não editar) de zero-value (ex.: active=false, preço=0) —
// necessário porque o PATCH é parcial.
type planRequest struct {
	Name            *string   `json:"name"`
	Slug            *string   `json:"slug"`
	Type            *string   `json:"type"`
	PriceMonthly    *int      `json:"price_monthly"`
	PriceYearly     *int      `json:"price_yearly"`
	MaxJobs              *int      `json:"max_jobs"`
	MaxResumeAnalyses    *int      `json:"max_resume_analyses"`
	MaxSites             *int      `json:"max_sites"`
	MaxCrawlsPerDay *int      `json:"max_crawls_per_day"`
	AIMatching      *bool     `json:"ai_matching"`
	InterviewPrep   *bool     `json:"interview_prep"`
	Features        *[]string `json:"features"`
	Active          *bool     `json:"active"`
	IsDefault       *bool     `json:"is_default"`
}

// ListPlansAdmin lista todos os planos (inclusive inativos) para a gestão.
func ListPlansAdmin(c *gin.Context) {
	var plans []models.Plan
	if err := DB.Order("id").Find(&plans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar planos"})
		return
	}
	if plans == nil {
		plans = []models.Plan{}
	}
	c.JSON(http.StatusOK, plans)
}

// CreatePlan cria um plano. Novos planos nascem ativos e nunca como padrão —
// o padrão é atribuído explicitamente via PATCH (contract §2, V4–V6).
func CreatePlan(c *gin.Context) {
	var body planRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := strings.TrimSpace(strOr(body.Name, ""))
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nome é obrigatório"})
		return
	}
	slug := strings.TrimSpace(strOr(body.Slug, ""))
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Identificador é obrigatório"})
		return
	}
	if !slugPattern.MatchString(slug) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Identificador inválido: use letras minúsculas, números e hífen"})
		return
	}

	planType, ok := parsePlanType(strOr(body.Type, string(models.PlanTypeFree)))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tipo de plano inválido: use 'free' ou 'paid'"})
		return
	}

	priceMonthly := intOr(body.PriceMonthly, 0)
	priceYearly := intOr(body.PriceYearly, 0)
	if msg := validatePlanPrice(planType, priceMonthly, priceYearly); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	features := "[]"
	if body.Features != nil {
		encoded, err := json.Marshal(*body.Features)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Recursos inválidos"})
			return
		}
		features = string(encoded)
	}

	var existing int64
	DB.Model(&models.Plan{}).Where("slug = ?", slug).Count(&existing)
	if existing > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Identificador já existe"})
		return
	}

	plan := models.Plan{
		Name:            name,
		Slug:            slug,
		Type:            planType,
		PriceMonthly:    priceMonthly,
		PriceYearly:     priceYearly,
		MaxJobs:         intOr(body.MaxJobs, 100),
		MaxResumeAnalyses: intOr(body.MaxResumeAnalyses, 3),
		MaxSites:        intOr(body.MaxSites, 5),
		MaxCrawlsPerDay: intOr(body.MaxCrawlsPerDay, 10),
		AIMatching:      boolOr(body.AIMatching, false),
		InterviewPrep:   boolOr(body.InterviewPrep, false),
		Features:        features,
		Active:          true,
	}
	if err := DB.Create(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar plano"})
		return
	}
	c.JSON(http.StatusCreated, plan)
}

// UpdatePlan edita um plano existente (PATCH parcial). O slug é imutável e as
// regras do padrão de cadastro são validadas antes de qualquer escrita (V7–V9).
func UpdatePlan(c *gin.Context) {
	id, err := parseUintParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Identificador de plano inválido"})
		return
	}

	var plan models.Plan
	if err := DB.First(&plan, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plano não encontrado"})
		return
	}

	var body planRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if body.Slug != nil && strings.TrimSpace(*body.Slug) != plan.Slug {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Identificador não pode ser alterado"})
		return
	}

	// Estado resultante após aplicar o payload sobre o plano atual.
	newType := plan.Type
	if body.Type != nil {
		parsed, ok := parsePlanType(*body.Type)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tipo de plano inválido: use 'free' ou 'paid'"})
			return
		}
		newType = parsed
	}
	newMonthly := intOr(body.PriceMonthly, plan.PriceMonthly)
	newYearly := intOr(body.PriceYearly, plan.PriceYearly)
	newActive := plan.Active
	if body.Active != nil {
		newActive = *body.Active
	}

	name := plan.Name
	if body.Name != nil {
		name = strings.TrimSpace(*body.Name)
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Nome é obrigatório"})
			return
		}
	}

	if msg := validatePlanPrice(newType, newMonthly, newYearly); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	// Regras do padrão de cadastro (V7–V9).
	if plan.IsDefault {
		if !newActive || newType != models.PlanTypeFree {
			c.JSON(http.StatusConflict, gin.H{"error": "Plano é o padrão de cadastro; marque outro plano gratuito ativo como padrão antes"})
			return
		}
	}
	if body.IsDefault != nil && *body.IsDefault && !plan.IsDefault {
		if newType != models.PlanTypeFree || !newActive {
			c.JSON(http.StatusConflict, gin.H{"error": "Padrão de cadastro deve ser um plano gratuito ativo"})
			return
		}
	}

	updates := map[string]interface{}{
		"name":               name,
		"type":               newType,
		"price_monthly":      newMonthly,
		"price_yearly":       newYearly,
		"max_jobs":           intOr(body.MaxJobs, plan.MaxJobs),
		"max_resume_analyses": intOr(body.MaxResumeAnalyses, plan.MaxResumeAnalyses),
		"max_sites":          intOr(body.MaxSites, plan.MaxSites),
		"max_crawls_per_day": intOr(body.MaxCrawlsPerDay, plan.MaxCrawlsPerDay),
		"ai_matching":        boolOr(body.AIMatching, plan.AIMatching),
		"interview_prep":     boolOr(body.InterviewPrep, plan.InterviewPrep),
		"active":             newActive,
	}
	if body.Features != nil {
		encoded, err := json.Marshal(*body.Features)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Recursos inválidos"})
			return
		}
		updates["features"] = string(encoded)
	}
	if body.IsDefault != nil {
		updates["is_default"] = *body.IsDefault
	}

	err = DB.Transaction(func(tx *gorm.DB) error {
		if body.IsDefault != nil && *body.IsDefault && !plan.IsDefault {
			if err := tx.Model(&models.Plan{}).Where("is_default = ?", true).
				Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return tx.Model(&models.Plan{}).Where("id = ?", plan.ID).Updates(updates).Error
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar plano"})
		return
	}

	var updated models.Plan
	if err := DB.First(&updated, plan.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao carregar plano"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func parsePlanType(raw string) (models.PlanType, bool) {
	switch models.PlanType(strings.TrimSpace(raw)) {
	case models.PlanTypeFree:
		return models.PlanTypeFree, true
	case models.PlanTypePaid:
		return models.PlanTypePaid, true
	default:
		return "", false
	}
}

// validatePlanPrice aplica V2/V3 ao estado resultante do plano.
func validatePlanPrice(t models.PlanType, monthly, yearly int) string {
	if monthly < 0 || yearly < 0 {
		return "Preço não pode ser negativo"
	}
	if t == models.PlanTypeFree && (monthly != 0 || yearly != 0) {
		return "Plano gratuito deve ter preço zero"
	}
	if t == models.PlanTypePaid && monthly <= 0 && yearly <= 0 {
		return "Plano pago exige preço maior que zero"
	}
	return ""
}

func intOr(v *int, fallback int) int {
	if v == nil {
		return fallback
	}
	return *v
}

func boolOr(v *bool, fallback bool) bool {
	if v == nil {
		return fallback
	}
	return *v
}

func strOr(v *string, fallback string) string {
	if v == nil {
		return fallback
	}
	return *v
}
