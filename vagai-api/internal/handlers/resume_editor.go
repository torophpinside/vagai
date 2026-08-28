package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/anomalyco/vagai-api/internal/models"
	"github.com/anomalyco/vagai-api/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ParseResume(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arquivo nao enviado"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".pdf" && ext != ".docx" && ext != ".txt" && ext != ".doc" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tipo de arquivo nao suportado. Use PDF, DOCX ou TXT"})
		return
	}

	tmpDir := "./uploads/resumes/tmp"
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao criar diretorio temporario"})
		return
	}

	filePath := filepath.Join(tmpDir, fmt.Sprintf("parse_%d_%s", time.Now().UnixNano(), file.Filename))
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao salvar arquivo"})
		return
	}
	defer os.Remove(filePath)

	rawText, err := services.ExtractTextFromFile(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao extrair texto do arquivo"})
		return
	}

	if rawText == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nao foi possivel extrair texto do arquivo"})
		return
	}

	resumeData, err := services.ParseResumeFields(rawText)
	if err != nil {
		log.Printf("Erro no parse AI, usando fallback: %v", err)
	}

	dataJSON, err := json.Marshal(resumeData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar dados estruturados"})
		return
	}

	allowed, current, maxLimit, err := checkPlanLimit(db, orgID, resourceResumes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar limite do plano"})
		return
	}
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   fmt.Sprintf("Limite de curriculos atingido: %d/%d", current, maxLimit),
			"current": current,
			"limit":   maxLimit,
		})
		return
	}

	resume := models.Resume{
		OrganizationID: orgID,
		Name:           file.Filename,
		FilePath:       filePath,
		Content:        rawText,
		Data:           string(dataJSON),
		Version:        1,
		UploadedAt:     time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := db.Create(&resume).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar curriculo"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Curriculo parseado com sucesso",
		"resume":  resume,
		"data":    resumeData,
	})
}

func GetResumeData(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalido"})
		return
	}

	var resume models.Resume
	if err := db.Where("id = ? AND organization_id = ?", id, orgID).First(&resume).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Curriculo nao encontrado"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar curriculo"})
		}
		return
	}

	if resume.Data == "" {
		c.JSON(http.StatusOK, gin.H{
			"resume_id": resume.ID,
			"name":      resume.Name,
			"data":      services.ResumeData{},
			"version":   resume.Version,
		})
		return
	}

	var resumeData services.ResumeData
	if err := json.Unmarshal([]byte(resume.Data), &resumeData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao decodificar dados do curriculo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"resume_id": resume.ID,
		"name":      resume.Name,
		"data":      resumeData,
		"version":   resume.Version,
	})
}

func UpdateResumeData(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalido"})
		return
	}

	var resume models.Resume
	if err := db.Where("id = ? AND organization_id = ?", id, orgID).First(&resume).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Curriculo nao encontrado"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar curriculo"})
		}
		return
	}

	var resumeData services.ResumeData
	if err := c.ShouldBindJSON(&resumeData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados invalidos: " + err.Error()})
		return
	}

	dataJSON, err := json.Marshal(resumeData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao serializar dados"})
		return
	}

	resume.Data = string(dataJSON)
	resume.Version++
	resume.UpdatedAt = time.Now()

	if err := db.Save(&resume).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar dados"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Dados salvos com sucesso",
		"resume":  resume,
		"data":    resumeData,
	})
}

func GenerateResumePDFHandler(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalido"})
		return
	}

	var resume models.Resume
	if err := db.Where("id = ? AND organization_id = ?", id, orgID).First(&resume).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Curriculo nao encontrado"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar curriculo"})
		}
		return
	}

	if resume.Data == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Curriculo nao possui dados estruturados. Faca o parse primeiro."})
		return
	}

	var resumeData services.ResumeData
	if err := json.Unmarshal([]byte(resume.Data), &resumeData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao decodificar dados do curriculo"})
		return
	}

	pdfBytes, err := services.GenerateResumePDF(resumeData)
	if err != nil {
		log.Printf("Erro ao gerar PDF: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar PDF"})
		return
	}

	fileName := "curriculo.pdf"
	if resumeData.PersonalInfo.Name != "" {
		name := strings.ReplaceAll(resumeData.PersonalInfo.Name, " ", "_")
		name = strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
				return r
			}
			return -1
		}, name)
		if name != "" {
			fileName = "curriculo_" + name + ".pdf"
		}
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
