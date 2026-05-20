package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"strings"

	"github.com/anomalyco/vagai-api/internal/middleware"
	"github.com/anomalyco/vagai-api/internal/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *gin.Context) {
	var body struct {
		Name         string `json:"name" binding:"required,min=2,max=200"`
		Email        string `json:"email" binding:"required,email"`
		Password     string `json:"password" binding:"required,min=8"`
		Organization string `json:"organization" binding:"required,min=2,max=200"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingUser models.User
	if err := DB.Where("email = ?", strings.ToLower(body.Email)).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email já cadastrado"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno"})
		return
	}

	tx := DB.Begin()
	defer tx.Rollback()

	org := models.Organization{
		Name: body.Organization,
		Slug: generateSlug(body.Organization),
		Plan: "free",
	}
	if err := tx.Create(&org).Error; err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			org.Slug = org.Slug + "-" + randomString(4)
			if err := tx.Create(&org).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar organização"})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar organização"})
			return
		}
	}

	user := models.User{
		Name:         body.Name,
		Email:        strings.ToLower(body.Email),
		PasswordHash: string(hashedPassword),
	}
	if err := tx.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar usuário"})
		return
	}

	membership := models.Membership{
		OrganizationID: org.ID,
		UserID:         user.ID,
		Role:           models.RoleOwner,
	}
	if err := tx.Create(&membership).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar membership"})
		return
	}

	var freePlan models.Plan
	if err := tx.Where("slug = ?", "free").First(&freePlan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao configurar plano"})
		return
	}

	subscription := models.Subscription{
		OrganizationID: org.ID,
		PlanID:         freePlan.ID,
		Status:         models.SubStatusActive,
	}
	if err := tx.Create(&subscription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar assinatura"})
		return
	}

	tx.Commit()

	token, err := middleware.GenerateToken(user.ID, org.ID, user.Email, string(membership.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Conta criada com sucesso",
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  membership.Role,
		},
		"organization": gin.H{
			"id":   org.ID,
			"name": org.Name,
			"slug": org.Slug,
			"plan": org.Plan,
		},
		"token": token,
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

	var membership models.Membership
	if err := DB.Where("user_id = ?", user.ID).Preload("Organization").First(&membership).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário sem organização"})
		return
	}

	token, err := middleware.GenerateToken(user.ID, membership.OrganizationID, user.Email, string(membership.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login realizado com sucesso",
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  membership.Role,
		},
		"organization": gin.H{
			"id":   membership.Organization.ID,
			"name": membership.Organization.Name,
			"slug": membership.Organization.Slug,
			"plan": membership.Organization.Plan,
		},
		"token": token,
	})
}

func GetMe(c *gin.Context) {
	userID, _ := c.Get("user_id")
	orgID, _ := c.Get("org_id")

	var user models.User
	if err := DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	}

	var org models.Organization
	if err := DB.First(&org, orgID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organização não encontrada"})
		return
	}

	var membership models.Membership
	DB.Where("user_id = ? AND organization_id = ?", userID, orgID).First(&membership)

	var sub models.Subscription
	DB.Where("organization_id = ?", orgID).Order("created_at DESC").First(&sub)

	var plan models.Plan
	DB.Where("slug = ?", org.Plan).First(&plan)

	var jobCount, resumeCount, siteCount int64
	DB.Model(&models.Job{}).Where("organization_id = ?", orgID).Count(&jobCount)
	DB.Model(&models.Resume{}).Where("organization_id = ?", orgID).Count(&resumeCount)
	DB.Model(&models.Site{}).Where("organization_id = ?", orgID).Count(&siteCount)

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"avatar": user.Avatar,
			"timezone": user.Timezone,
			"role": membership.Role,
		},
		"organization": gin.H{
			"id":            org.ID,
			"name":          org.Name,
			"slug":          org.Slug,
			"plan":          org.Plan,
			"trial_ends_at": org.TrialEndsAt,
		},
		"plan": gin.H{
			"name":             plan.Name,
			"slug":             plan.Slug,
			"price_monthly":    plan.PriceMonthly,
			"price_yearly":     plan.PriceYearly,
			"max_jobs":         plan.MaxJobs,
			"max_resumes":      plan.MaxResumes,
			"max_sites":        plan.MaxSites,
			"features":         plan.Features,
		},
		"usage": gin.H{
			"jobs":    jobCount,
			"resumes": resumeCount,
			"sites":   siteCount,
		},
		"subscription": gin.H{
			"status":              sub.Status,
			"current_period_end":  sub.CurrentPeriodEnd,
			"cancel_at_period_end": sub.CancelAtPeriodEnd,
		},
	})
}

func UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var body struct {
		Name     string `json:"name" binding:"omitempty,min=2,max=200"`
		Timezone string `json:"timezone" binding:"omitempty"`
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

	if len(updates) > 0 {
		DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Perfil atualizado"})
}

func ChangePassword(c *gin.Context) {
	userID, _ := c.Get("user_id")

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
	orgID, _ := c.Get("org_id")

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

	var org models.Organization
	if err := DB.First(&org, orgID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organização não encontrada"})
		return
	}

	org.Plan = plan.Slug
	if err := DB.Save(&org).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar plano"})
		return
	}

	var sub models.Subscription
	if err := DB.Where("organization_id = ?", orgID).Order("created_at DESC").First(&sub).Error; err == nil {
		sub.PlanID = plan.ID
		DB.Save(&sub)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Plano alterado para " + plan.Name,
		"plan": gin.H{
			"name":           plan.Name,
			"slug":           plan.Slug,
			"price_monthly":  plan.PriceMonthly,
			"max_jobs":       plan.MaxJobs,
			"max_resumes":    plan.MaxResumes,
			"max_sites":      plan.MaxSites,
			"features":       plan.Features,
		},
	})
}

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, ".", "-")
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	return slug
}

func randomString(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)[:n]
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
