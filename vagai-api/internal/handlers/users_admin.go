package handlers

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/anomalyco/vagai-api/internal/models"
	"github.com/anomalyco/vagai-api/internal/services"
	"github.com/gin-gonic/gin"
)

// adminUserResponse inclui campos de gestão da conta excluída (janela de
// reativação) que o modelo esconde do JSON padrão.
type adminUserResponse struct {
	models.User
	PurgeAfterJSON *string `json:"purge_after"`
	DeletedAtJSON  *string `json:"deleted_at"`
}

func wrapAdminUser(u *models.User) adminUserResponse {
	resp := adminUserResponse{User: *u}
	if u.PurgeAfter != nil {
		s := u.PurgeAfter.Format("2006-01-02T15:04:05Z07:00")
		resp.PurgeAfterJSON = &s
	}
	if u.DeletedAt.Valid {
		s := u.DeletedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		resp.DeletedAtJSON = &s
	}
	return resp
}

// fetchManageableUser carrega uma conta do tipo User aplicando os guards de
// gestão (FR-002/FR-008): alvo inexistente → false (404), admin ou própria
// conta → false (403). Em falha o handler já respondeu.
func fetchManageableUser(c *gin.Context) (*models.User, bool) {
	id, err := parseUintParam(c.Param("id"))
	if err != nil {
		abortError(c, http.StatusBadRequest, "Identificador de usuário inválido")
		return nil, false
	}

	var user models.User
	if err := DB.Unscoped().First(&user, id).Error; err != nil {
		abortError(c, http.StatusNotFound, "Usuário não encontrado")
		return nil, false
	}
	if user.Role == models.RoleAdmin {
		abortError(c, http.StatusForbidden, "Contas administrativas não são gerenciáveis")
		return nil, false
	}
	if actorIDFrom(c) == user.ID {
		abortError(c, http.StatusForbidden, "Não é possível gerenciar a própria conta")
		return nil, false
	}
	return &user, true
}

// ListUsersAdmin lista contas do tipo User com busca, filtro de situação e
// paginação (FR-001/FR-003, contract §GET /api/admin/users).
func ListUsersAdmin(c *gin.Context) {
	status := strings.TrimSpace(c.DefaultQuery("status", "active"))
	query := DB.Unscoped().Where("role = ?", models.RoleUser)
	switch status {
	case "active":
		query = query.Where("suspended = ? AND deleted_at IS NULL", false)
	case "suspended":
		query = query.Where("suspended = ? AND deleted_at IS NULL", true)
	case "deleted":
		query = query.Where("deleted_at IS NOT NULL")
	default:
		abortError(c, http.StatusUnprocessableEntity, "Situação inválida: use 'active', 'suspended' ou 'deleted'")
		return
	}

	if search := strings.TrimSpace(c.Query("search")); search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ?", like, like)
	}

	var total int64
	if err := query.Model(&models.User{}).Count(&total).Error; err != nil {
		abortError(c, http.StatusInternalServerError, "Erro ao listar usuários")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var users []models.User
	if err := query.Order("id DESC").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		abortError(c, http.StatusInternalServerError, "Erro ao listar usuários")
		return
	}
	if users == nil {
		users = []models.User{}
	}

	data := make([]adminUserResponse, 0, len(users))
	for i := range users {
		data = append(data, wrapAdminUser(&users[i]))
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       data,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": (int(total) + limit - 1) / limit,
	})
}

// GetUserAdmin retorna o perfil de gestão com contadores de atividade
// (FR-004, contract §GET /api/admin/users/:id).
func GetUserAdmin(c *gin.Context) {
	user, ok := fetchManageableUser(c)
	if !ok {
		return
	}

	// Vagas e matches = totais; candidaturas = matches applied=true;
	// currículos = análises criadas no mês corrente (mesmo critério do
	// usage em auth.go).
	var jobs, resumeAnalyses, matches, applied int64
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	DB.Model(&models.Job{}).Where("user_id = ?", user.ID).Count(&jobs)
	DB.Model(&models.ResumeAnalysis{}).Where("user_id = ? AND created_at >= ?", user.ID, monthStart).Count(&resumeAnalyses)
	DB.Model(&models.Match{}).Where("user_id = ?", user.ID).Count(&matches)
	DB.Model(&models.Match{}).Where("user_id = ? AND applied = ?", user.ID, true).Count(&applied)

	var deletedAt, purgeAfter interface{}
	if user.DeletedAt.Valid {
		deletedAt = user.DeletedAt.Time
	}
	if user.PurgeAfter != nil {
		purgeAfter = *user.PurgeAfter
	}

	c.JSON(http.StatusOK, gin.H{
		"id":                   user.ID,
		"name":                 user.Name,
		"email":                user.Email,
		"city":                 user.City,
		"timezone":             user.Timezone,
		"avatar":               user.Avatar,
		"role":                 user.Role,
		"plan_slug":            user.PlanSlug,
		"subscription_status":  user.SubscriptionStatus,
		"trial_ends_at":        user.TrialEndsAt,
		"subscription_ends_at": user.SubscriptionEndsAt,
		"suspended":            user.Suspended,
		"email_verified_at":    user.EmailVerifiedAt,
		"created_at":           user.CreatedAt,
		"updated_at":           user.UpdatedAt,
		"deleted_at":           deletedAt,
		"purge_after":          purgeAfter,
		"activity": gin.H{
			"jobs":    jobs,
			"resumes": resumeAnalyses,
			"matches": matches,
			"applied": applied,
		},
	})
}

// adminUserRequest é o payload de edição básica. Ponteiros distinguem
// "campo ausente" de zero-value (PATCH parcial, padrão planRequest).
type adminUserRequest struct {
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	City     *string `json:"city"`
	Timezone *string `json:"timezone"`
}

// UpdateUserAdmin edita dados básicos (FR-004/FR-011).
func UpdateUserAdmin(c *gin.Context) {
	user, ok := fetchManageableUser(c)
	if !ok {
		return
	}

	var body adminUserRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		abortError(c, http.StatusBadRequest, err.Error())
		return
	}

	updates := map[string]interface{}{}
	before := map[string]interface{}{}

	if body.Name != nil {
		name := strings.TrimSpace(*body.Name)
		if name == "" {
			abortError(c, http.StatusUnprocessableEntity, "Nome é obrigatório")
			return
		}
		if name != user.Name {
			updates["name"] = name
			before["name"] = user.Name
		}
	}
	if body.Email != nil {
		email := strings.ToLower(strings.TrimSpace(*body.Email))
		if email == "" {
			abortError(c, http.StatusUnprocessableEntity, "E-mail é obrigatório")
			return
		}
		if email != strings.ToLower(user.Email) {
			var existing int64
			DB.Unscoped().Model(&models.User{}).Where("LOWER(email) = ? AND id != ?", email, user.ID).Count(&existing)
			if existing > 0 {
				abortError(c, http.StatusConflict, "E-mail já cadastrado")
				return
			}
			updates["email"] = email
			before["email"] = user.Email
		}
	}
	if body.City != nil {
		updates["city"] = strings.TrimSpace(*body.City)
	}
	if body.Timezone != nil {
		updates["timezone"] = strings.TrimSpace(*body.Timezone)
	}

	if len(updates) == 0 {
		c.JSON(http.StatusOK, wrapAdminUser(user))
		return
	}

	if err := DB.Model(&models.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		abortError(c, http.StatusInternalServerError, "Erro ao atualizar usuário")
		return
	}
	logAudit(c, actorIDFrom(c), user.ID, "user.update", user.ID, gin.H{"before": before, "fields": keysOf(updates)})

	var updated models.User
	if err := DB.Unscoped().First(&updated, user.ID).Error; err != nil {
		abortError(c, http.StatusInternalServerError, "Erro ao carregar usuário")
		return
	}
	c.JSON(http.StatusOK, wrapAdminUser(&updated))
}

type assignPlanRequest struct {
	PlanSlug *string `json:"plan_slug"`
}

// AssignPlanAdmin atribui um plano ativo (FR-005, clarificação Q2: ajuste
// de acesso, sem tocar assinatura cobrada).
func AssignPlanAdmin(c *gin.Context) {
	user, ok := fetchManageableUser(c)
	if !ok {
		return
	}

	var body assignPlanRequest
	if err := c.ShouldBindJSON(&body); err != nil || body.PlanSlug == nil || strings.TrimSpace(*body.PlanSlug) == "" {
		abortError(c, http.StatusBadRequest, "Plano é obrigatório")
		return
	}
	slug := strings.TrimSpace(*body.PlanSlug)

	var plan models.Plan
	if err := DB.Where("slug = ? AND active = ?", slug, true).First(&plan).Error; err != nil {
		abortError(c, http.StatusNotFound, "Plano não encontrado")
		return
	}
	if user.PlanSlug == slug {
		c.JSON(http.StatusOK, gin.H{"message": "Plano atribuído", "user": wrapAdminUser(user)})
		return
	}

	from := user.PlanSlug
	if err := DB.Model(&models.User{}).Where("id = ?", user.ID).Update("plan_slug", slug).Error; err != nil {
		abortError(c, http.StatusInternalServerError, "Erro ao atribuir plano")
		return
	}
	logAudit(c, actorIDFrom(c), user.ID, "user.plan_assign", user.ID, gin.H{"from": from, "to": slug})

	var updated models.User
	if err := DB.Unscoped().First(&updated, user.ID).Error; err != nil {
		abortError(c, http.StatusInternalServerError, "Erro ao carregar usuário")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Plano atribuído", "user": wrapAdminUser(&updated)})
}

// SuspendUserAdmin suspende a conta e bloqueia sessões abertas (FR-006/FR-010).
func SuspendUserAdmin(c *gin.Context) {
	user, ok := fetchManageableUser(c)
	if !ok {
		return
	}
	if user.DeletedAt.Valid {
		abortError(c, http.StatusNotFound, "Usuário não encontrado")
		return
	}
	if user.Suspended {
		c.JSON(http.StatusOK, gin.H{"message": "Usuário já está suspenso"})
		return
	}

	if err := DB.Model(&models.User{}).Where("id = ?", user.ID).Update("suspended", true).Error; err != nil {
		abortError(c, http.StatusInternalServerError, "Erro ao suspender usuário")
		return
	}
	services.SetSuspendedMarker(user.ID)
	logAudit(c, actorIDFrom(c), user.ID, "user.suspend", user.ID, nil)
	c.JSON(http.StatusOK, gin.H{"message": "Usuário suspenso"})
}

// RestoreUserAdmin desfaz a suspensão (FR-006).
func RestoreUserAdmin(c *gin.Context) {
	user, ok := fetchManageableUser(c)
	if !ok {
		return
	}
	if user.DeletedAt.Valid {
		abortError(c, http.StatusNotFound, "Usuário não encontrado")
		return
	}
	if !user.Suspended {
		c.JSON(http.StatusOK, gin.H{"message": "Usuário já está ativo"})
		return
	}

	if err := DB.Model(&models.User{}).Where("id = ?", user.ID).Update("suspended", false).Error; err != nil {
		abortError(c, http.StatusInternalServerError, "Erro ao restaurar usuário")
		return
	}
	services.ClearSuspendedMarker(user.ID)
	logAudit(c, actorIDFrom(c), user.ID, "user.restore", user.ID, nil)
	c.JSON(http.StatusOK, gin.H{"message": "Usuário restaurado"})
}

// DeleteUserAdmin exclui a conta em duas fases (FR-007): soft-delete +
// purge_after; janela de reativação; e-mail reservado pelo uniqueIndex.
func DeleteUserAdmin(c *gin.Context) {
	user, ok := fetchManageableUser(c)
	if !ok {
		return
	}
	if user.DeletedAt.Valid {
		abortError(c, http.StatusConflict, "Usuário já excluído")
		return
	}

	purgeAfter := time.Now().Add(userDeleteRetention())
	if err := DB.Model(&models.User{}).Where("id = ?", user.ID).
		Updates(map[string]interface{}{
			"deleted_at":  time.Now(),
			"purge_after": purgeAfter,
			"suspended":   true,
		}).Error; err != nil {
		abortError(c, http.StatusInternalServerError, "Erro ao excluir usuário")
		return
	}
	services.SetSuspendedMarker(user.ID)
	logAudit(c, actorIDFrom(c), user.ID, "user.delete", user.ID, gin.H{"purge_after": purgeAfter})

	c.JSON(http.StatusOK, gin.H{"message": "Usuário excluído; dados serão removidos em " + retentionDaysLabel()})
}

// ReactivateUserAdmin reativa conta na janela (FR-007, SC-009).
func ReactivateUserAdmin(c *gin.Context) {
	user, ok := fetchManageableUser(c)
	if !ok {
		return
	}
	if !user.DeletedAt.Valid {
		abortError(c, http.StatusConflict, "Usuário não está excluído")
		return
	}
	if user.PurgeAfter != nil && time.Now().After(*user.PurgeAfter) {
		abortError(c, http.StatusConflict, "Janela de reativação expirada")
		return
	}

	if err := DB.Unscoped().Model(&models.User{}).Where("id = ?", user.ID).
		Updates(map[string]interface{}{
			"deleted_at":  nil,
			"purge_after": nil,
			"suspended":   false,
		}).Error; err != nil {
		abortError(c, http.StatusInternalServerError, "Erro ao reativar usuário")
		return
	}
	services.ClearSuspendedMarker(user.ID)
	logAudit(c, actorIDFrom(c), user.ID, "user.reactivate", user.ID, nil)

	var updated models.User
	if err := DB.Unscoped().First(&updated, user.ID).Error; err != nil {
		abortError(c, http.StatusInternalServerError, "Erro ao carregar usuário")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Usuário reativado", "user": wrapAdminUser(&updated)})
}

// userDeleteRetention lê a janela de reativação (default 30 dias,
// research.md R3; env para operação e teste acelerado).
func userDeleteRetention() time.Duration {
	days := 30
	if raw := strings.TrimSpace(os.Getenv("USER_DELETE_RETENTION_DAYS")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n >= 0 {
			days = n
		}
	}
	return time.Duration(days) * 24 * time.Hour
}

func retentionDaysLabel() string {
	days := int(userDeleteRetention() / (24 * time.Hour))
	return strconv.Itoa(days) + " dias"
}

func keysOf(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
