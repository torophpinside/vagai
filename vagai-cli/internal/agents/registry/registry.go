package registry

import (
	"fmt"
	"log"

	"github.com/anomalyco/vagai-cli/internal/db"
	"github.com/anomalyco/vagai-cli/internal/models"
)

func AddSite(name, url, selector string) error {
	if err := db.Init(); err != nil {
		return fmt.Errorf("falha ao inicializar banco: %w", err)
	}

	site := models.Site{
		Name:          name,
		URL:           url,
		SelectorLinks: selector,
		Active:        true,
	}

	if err := db.DB.Create(&site).Error; err != nil {
		return fmt.Errorf("falha ao adicionar site: %w", err)
	}

	log.Printf("Site adicionado: %s (%s)", name, url)
	db.Log(0, "registry", "site_added", fmt.Sprintf(`{"name": "%s", "url": "%s"}`, name, url))

	return nil
}

func ListSites() ([]models.Site, error) {
	if err := db.Init(); err != nil {
		return nil, fmt.Errorf("falha ao inicializar banco: %w", err)
	}

	var sites []models.Site
	db.DB.Find(&sites)
	return sites, nil
}

func GetActiveSites() ([]models.Site, error) {
	if err := db.Init(); err != nil {
		return nil, fmt.Errorf("falha ao inicializar banco: %w", err)
	}

	var sites []models.Site
	db.DB.Where("active = ?", true).Find(&sites)
	return sites, nil
}
