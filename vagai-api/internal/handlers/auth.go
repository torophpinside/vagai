package handlers

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/anomalyco/vagai-api/internal/middleware"
	"github.com/anomalyco/vagai-api/internal/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *gin.Context) {
	var body struct {
		Name     string `json:"name" binding:"required,min=2,max=200"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	email := strings.ToLower(body.Email)
	var existingUser models.User
	if err := DB.Where("email = ?", email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email já cadastrado"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno"})
		return
	}

	role := models.RoleUser
	if middleware.IsAdminEmail(email) {
		role = models.RoleAdmin
	}

	trialEnds := time.Now().Add(14 * 24 * time.Hour)
	user := models.User{
		Name:               body.Name,
		Email:              email,
		PasswordHash:       string(hashedPassword),
		Role:               role,
		PlanSlug:           "free",
		SubscriptionStatus: models.SubStatusTrial,
		TrialEndsAt:        &trialEnds,
	}
	if err := DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar usuário"})
		return
	}

	token, err := middleware.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Conta criada com sucesso",
		"user":    userResponse(user),
		"token":   token,
	})
}

func Login(c *gin.Context) {
	var body struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := DB.Where("email = ?", strings.ToLower(body.Email)).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email ou senha incorretos"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email ou senha incorretos"})
		return
	}

	if user.Suspended {
		c.JSON(http.StatusForbidden, gin.H{"error": "Conta suspensa"})
		return
	}

	token, err := middleware.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login realizado com sucesso",
		"user":    userResponse(user),
		"token":   token,
	})
}

// userResponse serializa os dados públicos do usuário para o frontend.
func userResponse(user models.User) gin.H {
	return gin.H{
		"id":                  user.ID,
		"name":                user.Name,
		"email":               user.Email,
		"avatar":              user.Avatar,
		"timezone":            user.Timezone,
		"city":                user.City,
		"role":                user.Role,
		"plan":                user.PlanSlug,
		"plan_slug":           user.PlanSlug,
		"subscription_status": user.SubscriptionStatus,
		"subscription_ends_at": user.SubscriptionEndsAt,
		"trial_ends_at":       user.TrialEndsAt,
	}
}

func GetMe(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user models.User
	if err := DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	}

	var plan models.Plan
	DB.Where("slug = ?", user.PlanSlug).First(&plan)

	var jobCount, resumeCount, siteCount int64
	DB.Model(&models.Job{}).Where("user_id = ?", userID).Count(&jobCount)
	DB.Model(&models.Resume{}).Where("user_id = ?", userID).Count(&resumeCount)
	DB.Model(&models.Site{}).Where("user_id = ?", userID).Count(&siteCount)

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":       user.ID,
			"name":     user.Name,
			"email":    user.Email,
			"avatar":   user.Avatar,
			"timezone": user.Timezone,
			"city":     user.City,
			"role":     user.Role,
			"plan":     user.PlanSlug,
		},
		"plan": gin.H{
			"name":          plan.Name,
			"slug":          plan.Slug,
			"price_monthly": plan.PriceMonthly,
			"price_yearly":  plan.PriceYearly,
			"max_jobs":      plan.MaxJobs,
			"max_resumes":   plan.MaxResumes,
			"max_sites":     plan.MaxSites,
			"features":      plan.Features,
		},
		"usage": gin.H{
			"jobs":    jobCount,
			"resumes": resumeCount,
			"sites":   siteCount,
		},
		"subscription": gin.H{
			"status":             user.SubscriptionStatus,
			"current_period_end": user.SubscriptionEndsAt,
			"cancel_at_period_end": false,
		},
	})
}

func UpdateProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	var body struct {
		Name     string `json:"name" binding:"omitempty,min=2,max=200"`
		Timezone string `json:"timezone" binding:"omitempty"`
		City     string `json:"city" binding:"omitempty,max=100"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if body.Name != "" {
		updates["name"] = body.Name
	}
	if body.Timezone != "" {
		updates["timezone"] = body.Timezone
	}
	if body.City != "" {
		updates["city"] = body.City
	}

	if len(updates) > 0 {
		DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates)
	}

	var user models.User
	DB.First(&user, userID)

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":       user.ID,
			"name":     user.Name,
			"email":    user.Email,
			"avatar":   user.Avatar,
			"timezone": user.Timezone,
			"city":     user.City,
		},
	})
}

func ChangePassword(c *gin.Context) {
	userID := c.GetUint("user_id")

	var body struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Senha atual incorreta"})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
	DB.Model(&user).Update("password_hash", string(hashedPassword))

	c.JSON(http.StatusOK, gin.H{"message": "Senha alterada com sucesso"})
}

func ChangePlan(c *gin.Context) {
	userID := c.GetUint("user_id")

	var body struct {
		PlanSlug string `json:"plan_slug" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var plan models.Plan
	if err := DB.Where("slug = ?", body.PlanSlug).First(&plan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plano não encontrado"})
		return
	}

	var user models.User
	if err := DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	}

	// Paywall: upgrade para plano pago exige assinatura Stripe ativa. Sem isso,
	// qualquer usuário poderia se auto-promover a Pro sem pagar.
	if (plan.PriceMonthly > 0 || plan.PriceYearly > 0) && user.PlanSlug != plan.Slug {
		if user.StripeCustomerID == "" || user.SubscriptionStatus != models.SubStatusActive {
			c.JSON(http.StatusPaymentRequired, gin.H{"error": "Para ativar um plano pago, conclua a assinatura pelo checkout. Contate o suporte se você já pagou."})
			return
		}
	}

	user.PlanSlug = plan.Slug
	if err := DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar plano"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Plano alterado para " + plan.Name,
		"plan": gin.H{
			"name":          plan.Name,
			"slug":          plan.Slug,
			"price_monthly": plan.PriceMonthly,
			"max_jobs":      plan.MaxJobs,
			"max_resumes":   plan.MaxResumes,
			"max_sites":     plan.MaxSites,
			"features":      plan.Features,
		},
	})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
