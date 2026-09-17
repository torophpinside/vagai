package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/anomalyco/vagai-api/internal/models"
	"github.com/anomalyco/vagai-api/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var DB *gorm.DB

func SetDB(db *gorm.DB) {
	DB = db
}

// trackingParams são parâmetros de query que devem ser removidos na normalização
var trackingParams = map[string]bool{
	"utm_source": true, "utm_medium": true, "utm_campaign": true,
	"utm_term": true, "utm_content": true, "ref": true,
	"source": true, "fbclid": true, "gclid": true,
	"hsa_cam": true, "hsa_grp": true, "hsa_mt": true,
	"hsa_src": true, "hsa_ad": true, "hsa_acc": true,
	"hsa_net": true, "hsa_ver": true, "mc_cid": true,
	"mc_eid": true, "yclid": true, "msclkid": true,
}

// NormalizeURL normaliza URLs para evitar duplicatas por diferenças
// insignificantes (query params de tracking, trailing slashes, etc.)
func NormalizeURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return rawURL
	}

	if strings.HasPrefix(rawURL, "//") {
		rawURL = "https:" + rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	parsed.Scheme = "https"

	host := parsed.Hostname()
	if strings.HasPrefix(host, "www.") {
		host = host[4:]
	}
	if port := parsed.Port(); port != "" && port != "443" {
		host = host + ":" + port
	}
	parsed.Host = host

	q := parsed.Query()
	modified := false
	for key := range q {
		if trackingParams[strings.ToLower(key)] {
			q.Del(key)
			modified = true
		}
	}
	if modified {
		parsed.RawQuery = q.Encode()
	} else {
		parsed.RawQuery = ""
	}

	parsed.Fragment = ""

	path := parsed.Path
	if path != "/" && strings.HasSuffix(path, "/") {
		path = strings.TrimRight(path, "/")
	}
	parsed.Path = path

	return parsed.String()
}

func getDB(c *gin.Context) *gorm.DB {
	if scoped, exists := c.Get("scoped_db"); exists {
		return scoped.(*gorm.DB)
	}
	return DB
}

type planResource string

const (
	resourceSites   planResource = "sites"
	resourceResumes planResource = "resumes"
	resourceJobs    planResource = "jobs"
)

func checkPlanLimit(db *gorm.DB, orgID uint, resource planResource) (allowed bool, current int, maxLimit int, err error) {
	var org models.Organization
	if err := db.First(&org, orgID).Error; err != nil {
		return false, 0, 0, err
	}

	var plan models.Plan
	if err := db.Where("slug = ?", org.Plan).First(&plan).Error; err != nil {
		return false, 0, 0, err
	}

	switch resource {
	case resourceSites:
		maxLimit = plan.MaxSites
	case resourceResumes:
		maxLimit = plan.MaxResumes
	case resourceJobs:
		maxLimit = plan.MaxJobs
	}

	if maxLimit == -1 {
		return true, 0, 0, nil
	}

	var count int64
	switch resource {
	case resourceSites:
		db.Model(&models.Site{}).Where("organization_id = ?", orgID).Count(&count)
	case resourceResumes:
		db.Model(&models.Resume{}).Where("organization_id = ? AND file_path <> ''", orgID).Count(&count)
	case resourceJobs:
		db.Model(&models.Job{}).Where("organization_id = ?", orgID).Count(&count)
	}
	current = int(count)

	return current < maxLimit, current, maxLimit, nil
}

func ListJobs(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")

	query := db.Where("organization_id = ?", orgID)

	if status := c.Query("status"); status != "" {
		statuses := strings.Split(status, ",")
		query = query.Where("status IN ?", statuses)
	} else {
		query = query.Where("status NOT IN ?", []string{"ignored", "unmatched"})
	}

	if site := c.Query("site"); site != "" {
		siteIDs := strings.Split(site, ",")
		ids := make([]uint, 0, len(siteIDs))
		for _, s := range siteIDs {
			if id, err := strconv.ParseUint(strings.TrimSpace(s), 10, 32); err == nil {
				ids = append(ids, uint(id))
			}
		}
		if len(ids) > 0 {
			query = query.Where("site_id IN ?", ids)
		}
	}

	if search := c.Query("search"); search != "" {
		query = query.Where("LOWER(title) LIKE ?", "%"+strings.ToLower(search)+"%")
	}

	var total int64
	query.Model(&models.Job{}).Count(&total)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var jobs []models.Job
	query.Order("collected_at DESC").Preload("Site").Offset(offset).Limit(limit).Find(&jobs)

	if jobs == nil {
		jobs = []models.Job{}
	}

	scoreMap := make(map[uint]float64)
	if len(jobs) > 0 {
		ids := make([]uint, len(jobs))
		for i, j := range jobs {
			ids[i] = j.ID
		}
		var scores []struct {
			JobID uint    `gorm:"column:job_id"`
			Score float64 `gorm:"column:score"`
		}
		db.Model(&models.Match{}).
			Select("job_id, MAX(similarity_score) as score").
			Where("job_id IN ?", ids).
			Group("job_id").
			Find(&scores)
		for _, s := range scores {
			scoreMap[s.JobID] = s.Score
		}
	}

	type jobItem struct {
		models.Job
		MaxScore float64 `json:"max_score"`
	}

	data := make([]jobItem, len(jobs))
	for i, j := range jobs {
		data[i] = jobItem{Job: j, MaxScore: scoreMap[j.ID]}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       data,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": (int(total) + limit - 1) / limit,
	})
}

func ExtractJob(c *gin.Context) {
	jobURL := c.PostForm("url")
	if jobURL == "" {
		var body struct {
			URL string `json:"url"`
		}
		if err := c.ShouldBindJSON(&body); err != nil || body.URL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "URL é obrigatória"})
			return
		}
		jobURL = body.URL
	}

	jobURL = NormalizeURL(jobURL)

	data, err := services.ExtractJobFromURL(jobURL)
	if err != nil {
		log.Printf("Erro ao extrair vaga de %s: %v", jobURL, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Não foi possível extrair a vaga desta URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"title":       data["title"],
		"company":     data["company"],
		"description": data["description"],
		"url":         jobURL,
	})
}

func CreateJob(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")

	var job models.Job
	if err := c.ShouldBindJSON(&job); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if job.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Título é obrigatório"})
		return
	}

	if job.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL é obrigatória"})
		return
	}

	job.URL = NormalizeURL(job.URL)

	allowed, current, maxLimit, err := checkPlanLimit(db, orgID, resourceJobs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar limite do plano"})
		return
	}
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   fmt.Sprintf("Limite de vagas atingido: %d/%d", current, maxLimit),
			"current": current,
			"limit":   maxLimit,
		})
		return
	}

	var existing models.Job
	if err := db.Where("url = ? AND organization_id = ?", job.URL, orgID).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Vaga já cadastrada com esta URL"})
		return
	}

	job.OrganizationID = orgID
	job.Status = models.JobStatusNew

	if err := db.Create(&job).Error; err != nil {
		log.Printf("Erro ao salvar vaga: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar vaga"})
		return
	}

	c.JSON(http.StatusCreated, job)
}

func GetJob(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var job models.Job
	if err := db.Where("id = ? AND organization_id = ?", id, orgID).Preload("Site").First(&job).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vaga não encontrada"})
		return
	}

	var matches []models.Match
	db.Where("job_id = ?", id).Preload("Resume", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Find(&matches)

	c.JSON(http.StatusOK, gin.H{"job": job, "matches": matches})
}

func UpdateJobStatus(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var job models.Job
	if err := db.Where("id = ? AND organization_id = ?", id, orgID).First(&job).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vaga não encontrada"})
		return
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job.Status = models.JobStatus(body.Status)
	db.Save(&job)

	var org models.Organization
	if err := db.First(&org, orgID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar organização"})
		return
	}
	var negativeKeywords []string
	if org.NegativeKeywords != "" {
		if err := json.Unmarshal([]byte(org.NegativeKeywords), &negativeKeywords); err != nil {
			log.Printf("Erro ao decodificar negative_keywords: %v", err)
		}
	}

	if body.Status == "matched" {
		var resumes []models.Resume
		if err := db.Where("organization_id = ?", orgID).Find(&resumes).Error; err == nil {
			for i := range resumes {
				resume := &resumes[i]
				if resume.Data == "" && resume.Content == "" {
					continue
				}

				var resumeData services.ResumeData
				if resume.Data != "" {
					if err := json.Unmarshal([]byte(resume.Data), &resumeData); err != nil {
						continue
					}
				}

				result := services.MatchResumeToJob(resumeData, resume.Content, job.Title, job.Description, negativeKeywords)
				if result.Score <= 0 {
					continue
				}

				keywordsJSON, _ := json.Marshal(result.KeywordsMatched)
				match := models.Match{
					OrganizationID:  orgID,
					JobID:           job.ID,
					ResumeID:        &resume.ID,
					SimilarityScore: result.Score,
					KeywordsMatched: string(keywordsJSON),
					AIReason: fmt.Sprintf(
						"Match calculado a partir dos dados do curriculo gravados no banco (%d termos-chave em comum).",
						len(result.KeywordsMatched),
					),
				}
				db.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "job_id"}, {Name: "resume_id"}},
					DoUpdates: clause.AssignmentColumns([]string{"similarity_score", "keywords_matched", "ai_reason"}),
				}).Create(&match)
			}
		}
	}

	c.JSON(http.StatusOK, job)
}

func RematchMatches(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")

	threshold := 65
	if t, err := strconv.Atoi(c.DefaultQuery("threshold", "65")); err == nil && t > 0 {
		threshold = t
	}

	var resume models.Resume
	if err := db.Where("organization_id = ?", orgID).
		Where("data <> '' OR content <> ''").
		Order("uploaded_at DESC").
		First(&resume).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Nenhum curriculo com dados salvo. Salve seu curriculo no editor antes de reanalisar."})
		return
	}

	var resumeData services.ResumeData
	if resume.Data != "" {
		if err := json.Unmarshal([]byte(resume.Data), &resumeData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao decodificar dados do curriculo"})
			return
		}
	}

	var org models.Organization
	if err := db.First(&org, orgID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar organização"})
		return
	}
	var negativeKeywords []string
	if org.NegativeKeywords != "" {
		if err := json.Unmarshal([]byte(org.NegativeKeywords), &negativeKeywords); err != nil {
			log.Printf("Erro ao decodificar negative_keywords: %v", err)
		}
	}

	var jobs []models.Job
	if err := db.Where("organization_id = ? AND status IN ?", orgID, []string{"new", "matched"}).
		Find(&jobs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar vagas"})
		return
	}

	var userCity string
	var user models.User
	if err := db.Where("organization_id = ? AND city <> ''", orgID).First(&user).Error; err == nil {
		userCity = user.City
	}

	created, updated, excluded := 0, 0, 0
	for _, job := range jobs {
		loc := services.AnalyzeJobLocation(job.Title + " " + job.Description)
		if !services.LocationAllowedToMatch(loc, userCity) {
			db.Model(&job).Update("status", models.JobStatusUnmatched)
			excluded++
			continue
		}

		result := services.MatchResumeToJob(resumeData, resume.Content, job.Title, job.Description, negativeKeywords)
		if result.Score <= 0 {
			continue
		}

		keywordsJSON, _ := json.Marshal(result.KeywordsMatched)
		match := models.Match{
			OrganizationID:  orgID,
			JobID:           job.ID,
			ResumeID:        &resume.ID,
			SimilarityScore: result.Score,
			KeywordsMatched: string(keywordsJSON),
			AIReason: fmt.Sprintf(
				"Match recalculado contra o curriculo mais recente (%d termos-chave em comum).",
				len(result.KeywordsMatched),
			),
		}
		res := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "job_id"}, {Name: "resume_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"similarity_score", "keywords_matched", "ai_reason"}),
		}).Create(&match)
		if res.Error != nil {
			continue
		}
		if res.RowsAffected == 0 {
			updated++
		} else {
			created++
		}

		newStatus := models.JobStatusAnalyzed
		if result.Score >= float64(threshold) {
			newStatus = models.JobStatusMatched
		}
		db.Model(&job).Update("status", newStatus)
	}

	c.JSON(http.StatusOK, gin.H{
		"resume_id":              resume.ID,
		"jobs_processed":         len(jobs),
		"matches_created":        created,
		"matches_updated":        updated,
		"jobs_excluded_location": excluded,
	})
}

func ListMatches(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")
	threshold, _ := strconv.Atoi(c.DefaultQuery("threshold", "65"))
	sortOrder := c.DefaultQuery("sort", "desc")
	appliedFilter := c.DefaultQuery("applied", "false")

	query := db.Model(&models.Match{}).
		Where("matches.organization_id = ?", orgID).
		Joins("JOIN jobs ON jobs.id = matches.job_id").
		Where("similarity_score >= ?", threshold)

	if appliedFilter == "true" {
		query = query.Where("matches.applied = ?", true).
			Where("jobs.status NOT IN ?", []string{"ignored", "unmatched"})
	} else {
		query = query.Where("matches.applied = ?", false).
			Where("jobs.status = ?", "matched")
	}

	if site := c.Query("site"); site != "" {
		siteIDs := strings.Split(site, ",")
		ids := make([]uint, 0, len(siteIDs))
		for _, s := range siteIDs {
			if id, err := strconv.ParseUint(strings.TrimSpace(s), 10, 32); err == nil {
				ids = append(ids, uint(id))
			}
		}
		if len(ids) > 0 {
			query = query.Where("jobs.site_id IN ?", ids)
		}
	}

	if sortOrder == "asc" {
		query = query.Order("similarity_score ASC")
	} else {
		query = query.Order("similarity_score DESC")
	}

	var total int64
	query.Count(&total)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var matches []models.Match
	query.Preload("Job", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "title", "company", "url", "site_id", "description")
	}).Preload("Resume", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Offset(offset).Limit(limit).Find(&matches)
	if matches == nil {
		matches = []models.Match{}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       matches,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": (int(total) + limit - 1) / limit,
	})
}

func UpdateMatch(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var match models.Match
	if err := db.Where("id = ? AND organization_id = ?", id, orgID).First(&match).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match não encontrado"})
		return
	}

	var body struct {
		Applied bool `json:"applied"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	match.Applied = body.Applied
	if body.Applied {
		now := time.Now()
		match.AppliedAt = &now
	} else {
		match.AppliedAt = nil
	}
	db.Save(&match)
	c.JSON(http.StatusOK, match)
}

func DeleteMatch(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var match models.Match
	if err := db.Where("id = ? AND organization_id = ?", id, orgID).First(&match).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match não encontrado"})
		return
	}

	if err := db.Delete(&match).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao deletar match"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Match deletado"})
}

func GetStats(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")

	var totalJobs, totalMatches, totalSites, totalApplied int64
	db.Model(&models.Job{}).Where("organization_id = ?", orgID).Count(&totalJobs)
	db.Model(&models.Match{}).Where("organization_id = ?", orgID).Count(&totalMatches)
	db.Model(&models.Match{}).Where("organization_id = ? AND applied = ?", orgID, true).Count(&totalApplied)
	db.Model(&models.Site{}).Where("organization_id = ? AND active = ?", orgID, true).Count(&totalSites)

	var jobs []models.Job
	db.Where("organization_id = ?", orgID).Find(&jobs)

	languageCount := make(map[string]int)
	keywordCount := make(map[string]int)

	commonLangs := []string{"python", "javascript", "typescript", "java", "go", "rust", "c++", "c#", "ruby", "php", "swift", "kotlin", "scala", "sql", "r"}
	commonKeywords := []string{"react", "vue", "angular", "node", "django", "flask", "spring", "docker", "kubernetes", "aws", "azure", "gcp", "linux", "mysql", "postgresql", "mongodb", "redis", "git", "devops", "agile", "scrum", "api", "rest", "graphql", "microservices", "machine learning", "ai", "data science"}

	for _, job := range jobs {
		textLower := strings.ToLower(job.Title + " " + job.Description)

		for _, lang := range commonLangs {
			pattern := `\b` + regexp.QuoteMeta(lang) + `\b`
			if regexp.MustCompile(pattern).MatchString(textLower) {
				languageCount[lang]++
			}
		}
		for _, kw := range commonKeywords {
			pattern := `\b` + regexp.QuoteMeta(kw) + `\b`
			if regexp.MustCompile(pattern).MatchString(textLower) {
				keywordCount[kw]++
			}
		}
	}

	sortedLanguages := []map[string]interface{}{}
	for lang, count := range languageCount {
		sortedLanguages = append(sortedLanguages, map[string]interface{}{"name": lang, "count": count})
	}
	sort.Slice(sortedLanguages, func(i, j int) bool {
		return sortedLanguages[i]["count"].(int) > sortedLanguages[j]["count"].(int)
	})
	if len(sortedLanguages) > 10 {
		sortedLanguages = sortedLanguages[:10]
	}

	sortedKeywords := []map[string]interface{}{}
	for kw, count := range keywordCount {
		sortedKeywords = append(sortedKeywords, map[string]interface{}{"name": kw, "count": count})
	}
	sort.Slice(sortedKeywords, func(i, j int) bool {
		return sortedKeywords[i]["count"].(int) > sortedKeywords[j]["count"].(int)
	})
	if len(sortedKeywords) > 10 {
		sortedKeywords = sortedKeywords[:10]
	}

	c.JSON(http.StatusOK, gin.H{
		"total_jobs":    totalJobs,
		"total_matches": totalMatches,
		"total_applied": totalApplied,
		"active_sites":  totalSites,
		"languages":     sortedLanguages,
		"keywords":      sortedKeywords,
	})
}

func ListSites(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")
	var sites []models.Site
	db.Where("organization_id = ?", orgID).Find(&sites)
	c.JSON(http.StatusOK, sites)
}

func DeleteSite(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var site models.Site
	if err := db.Where("id = ? AND organization_id = ?", id, orgID).First(&site).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Site não encontrado"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar site"})
		}
		return
	}

	// Deleta jobs associados primeiro
	if err := db.Where("site_id = ?", site.ID).Delete(&models.Job{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao remover jobs do site"})
		return
	}

	// Deleta o site
	if err := db.Delete(&site).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao remover site"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Site removido"})
}

func UpdateSite(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var site models.Site
	if err := db.Where("id = ? AND organization_id = ?", id, orgID).First(&site).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Site não encontrado"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar site"})
		}
		return
	}

	var body struct {
		Active *bool `json:"active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if body.Active != nil {
		site.Active = *body.Active
	}

	db.Save(&site)
	c.JSON(http.StatusOK, site)
}

func AddSite(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")

	var site models.Site
	if err := c.ShouldBindJSON(&site); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// A organização é sempre derivada do JWT autenticado: ignora qualquer valor
	// enviado no body, impedindo um usuário de criar um site em outra organização.
	site.OrganizationID = orgID

	allowed, current, maxLimit, err := checkPlanLimit(db, site.OrganizationID, resourceSites)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar limite do plano"})
		return
	}
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   fmt.Sprintf("Limite de sites atingido: %d/%d", current, maxLimit),
			"current": current,
			"limit":   maxLimit,
		})
		return
	}

	site.Active = true

	// Descobrir seletores automaticamente via IA
	log.Printf("Descobrindo seletores para: %s", site.URL)
	selectors, err := services.DiscoverSelectorsWithAI(site.URL)
	if err != nil {
		log.Printf("Erro ao descobrir seletores: %v", err)
		// Define valores padrão em caso de erro
		site.SelectorLinks = "a.job-link, a[href*='job']"
		site.SelectorCompany = ".company, .company-name"
		site.SelectorDescription = ".description, .job-description"
	} else {
		site.SelectorLinks = selectors["selector_links"]
		site.SelectorCompany = selectors["selector_company"]
		site.SelectorDescription = selectors["selector_description"]
		site.DelaySeconds = 2
		site.RespectRobots = true
	}

	log.Printf("Seletores descobertos: links=%s, company=%s, desc=%s",
		site.SelectorLinks, site.SelectorCompany, site.SelectorDescription)

	db.Create(&site)
	c.JSON(http.StatusCreated, site)
}

func ListResumes(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")
	var resumes []models.Resume
	db.Where("organization_id = ?", orgID).Find(&resumes)
	c.JSON(http.StatusOK, resumes)
}

func UploadResume(c *gin.Context) {
	db := getDB(c)
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arquivo não enviado"})
		return
	}

	orgID := c.GetUint("org_id")

	// Validação de upload (tamanho + magic bytes) antes de persistir.
	ufh, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao processar arquivo"})
		return
	}
	if err := validateUpload(file.Filename, file.Size, ufh); err != nil {
		ufh.Close()
		if errors.Is(err, errUploadTooLarge) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	ufh.Close()

	allowed, current, maxLimit, err := checkPlanLimit(db, orgID, resourceResumes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar limite do plano"})
		return
	}
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   fmt.Sprintf("Limite de currículos atingido: %d/%d", current, maxLimit),
			"current": current,
			"limit":   maxLimit,
		})
		return
	}

	uploadDir := "./uploads/resumes"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao criar diretório de uploads"})
		return
	}

	filePath := filepath.Join(uploadDir, time.Now().Format("20060102150405")+"_"+sanitizeFilename(file.Filename))
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao salvar arquivo"})
		return
	}

	log.Printf("Extraindo texto de: %s", filePath)
	content, err := services.ExtractTextFromFile(filePath)
	if err != nil {
		log.Printf("Erro na extração: %v", err)
	}

	var resume models.Resume
	resume.OrganizationID = orgID
	resume.Name = truncateName(file.Filename, 255)
	resume.FilePath = filePath
	resume.Content = content
	resume.UploadedAt = time.Now()

	if err := db.Create(&resume).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar no banco"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Currículo carregado e processado", "resume": resume})
}

func AnalyzeResume(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arquivo não enviado"})
		return
	}

	afh, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao processar arquivo"})
		return
	}
	if err := validateUpload(file.Filename, file.Size, afh); err != nil {
		afh.Close()
		if errors.Is(err, errUploadTooLarge) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	afh.Close()

	tmpDir := "./uploads/resumes"
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao criar diretório"})
		return
	}

	filePath := filepath.Join(tmpDir, "analysis_"+time.Now().Format("20060102150405")+"_"+sanitizeFilename(file.Filename))
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao salvar arquivo"})
		return
	}
	defer os.Remove(filePath)

	content, err := services.ExtractTextFromFile(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao extrair texto do arquivo"})
		return
	}

	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Não foi possível extrair texto do arquivo"})
		return
	}

	if len(content) > 8000 {
		content = content[:8000]
	}

	c.Set("RequestTimeout", 240*time.Second)

	analysisResult, aiErr := services.AnalyzeResumeWithAI(content)
	if aiErr != nil {
		log.Printf("Erro na análise de IA: %v", aiErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao analisar currículo"})
		return
	}

	strengths, _ := json.Marshal(toStringSlice(analysisResult["strengths"]))
	weaknesses, _ := json.Marshal(toStringSlice(analysisResult["weaknesses"]))
	suggestions, _ := json.Marshal(toStringSlice(analysisResult["suggestions"]))
	fullAnalysis, _ := analysisResult["fullAnalysis"].(string)

	analysis := models.ResumeAnalysis{
		OrganizationID: orgID,
		FileName:       truncateName(file.Filename, 255),
		FullAnalysis:   fullAnalysis,
		Strengths:      string(strengths),
		Weaknesses:     string(weaknesses),
		Suggestions:    string(suggestions),
	}

	if err := db.Create(&analysis).Error; err != nil {
		log.Printf("Erro ao salvar análise: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar análise"})
		return
	}

	log.Printf("Análise salva: ID=%d, File=%s", analysis.ID, analysis.FileName)

	var parsedStrengths, parsedWeaknesses, parsedSuggestions []string
	json.Unmarshal([]byte(analysis.Strengths), &parsedStrengths)
	json.Unmarshal([]byte(analysis.Weaknesses), &parsedWeaknesses)
	json.Unmarshal([]byte(analysis.Suggestions), &parsedSuggestions)

	c.JSON(http.StatusCreated, gin.H{
		"id":              analysis.ID,
		"organization_id": analysis.OrganizationID,
		"file_name":       analysis.FileName,
		"fullAnalysis":    analysis.FullAnalysis,
		"strengths":       parsedStrengths,
		"weaknesses":      parsedWeaknesses,
		"suggestions":     parsedSuggestions,
		"created_at":      analysis.CreatedAt,
	})
}

func DeleteResumeAnalysis(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var analysis models.ResumeAnalysis
	if err := db.Where("id = ? AND organization_id = ?", id, orgID).First(&analysis).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Análise não encontrada"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar análise"})
		}
		return
	}

	if err := db.Delete(&analysis).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao remover análise"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Análise removida"})
}

func formatAnalysisResponse(a models.ResumeAnalysis) gin.H {
	var strengths, weaknesses, suggestions []string
	json.Unmarshal([]byte(a.Strengths), &strengths)
	json.Unmarshal([]byte(a.Weaknesses), &weaknesses)
	json.Unmarshal([]byte(a.Suggestions), &suggestions)
	return gin.H{
		"id":              a.ID,
		"organization_id": a.OrganizationID,
		"file_name":       a.FileName,
		"fullAnalysis":    a.FullAnalysis,
		"strengths":       strengths,
		"weaknesses":      weaknesses,
		"suggestions":     suggestions,
		"created_at":      a.CreatedAt,
	}
}

func GetNegativeKeywords(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")

	var org models.Organization
	if err := db.First(&org, orgID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organização não encontrada"})
		return
	}

	var keywords []string
	if org.NegativeKeywords != "" {
		if err := json.Unmarshal([]byte(org.NegativeKeywords), &keywords); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao decodificar palavras-chave"})
			return
		}
	}

	c.JSON(http.StatusOK, keywords)
}

func UpdateNegativeKeywords(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")

	var body struct {
		Keywords []string `json:"keywords"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	keywordsJSON, err := json.Marshal(body.Keywords)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar palavras-chave"})
		return
	}

	if err := db.Model(&models.Organization{}).Where("id = ?", orgID).Update("negative_keywords", string(keywordsJSON)).Error; err != nil {
		log.Printf("Erro ao atualizar palavras-chave: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar palavras-chave"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Palavras-chave atualizadas com sucesso"})
}

func ListResumeAnalyses(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")
	var analyses []models.ResumeAnalysis
	db.Where("organization_id = ?", orgID).Order("created_at desc").Find(&analyses)
	result := make([]gin.H, len(analyses))
	for i, a := range analyses {
		result[i] = formatAnalysisResponse(a)
	}
	c.JSON(http.StatusOK, result)
}

func GetResumeAnalysis(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var analysis models.ResumeAnalysis
	if err := db.Where("id = ? AND organization_id = ?", id, orgID).First(&analysis).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Análise não encontrada"})
		return
	}
	c.JSON(http.StatusOK, formatAnalysisResponse(analysis))
}

func ListPlans(c *gin.Context) {
	var plans []models.Plan
	DB.Find(&plans)
	if plans == nil {
		plans = []models.Plan{}
	}
	c.JSON(http.StatusOK, plans)
}

func toStringSlice(v interface{}) []string {
	if v == nil {
		return nil
	}
	if s, ok := v.([]string); ok {
		return s
	}
	if arr, ok := v.([]interface{}); ok {
		result := make([]string, 0, len(arr))
		for _, item := range arr {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}
