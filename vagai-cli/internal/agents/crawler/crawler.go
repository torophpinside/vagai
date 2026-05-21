package crawler

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/anomalyco/vagai-cli/internal/db"
	"github.com/anomalyco/vagai-cli/internal/models"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
}

func Run(targetSite string) error {
	log.Println("Iniciando Crawler Agent...")

	if err := db.Init(); err != nil {
		return fmt.Errorf("falha ao inicializar banco: %w", err)
	}

	var sites []models.Site
	query := db.DB.Where("active = ?", true)
	if targetSite != "" {
		query = query.Where("name LIKE ?", "%"+targetSite+"%")
	}
	query.Find(&sites)

	if len(sites) == 0 {
		log.Println("Nenhum site ativo encontrado")
		return nil
	}

	for _, site := range sites {
		log.Printf("Processando site: %s", site.Name)
		if err := crawlSite(site); err != nil {
			log.Printf("Erro ao processar %s: %v", site.Name, err)
		}
	}

	log.Println("Crawler Agent finalizado")
	return nil
}

func crawlSite(site models.Site) error {
	log.Printf("Buscando vagas em: %s", site.URL)

	if site.OrganizationID > 0 {
		allowed, current, maxLimit, err := db.CheckJobsLimit(site.OrganizationID)
		if err != nil {
			log.Printf("Erro ao verificar limite do plano para org %d: %v", site.OrganizationID, err)
		} else if !allowed {
			log.Printf("Limite de vagas atingido para org %d: %d/%d. Pulando site %s", site.OrganizationID, current, maxLimit, site.Name)
			return nil
		}
	}

	var jobsFound int

	if site.Name == "remoteok" || strings.Contains(site.URL, "remoteok.com") {
		return crawlRemoteOKAPI(site)
	}

	if site.Name == "Working Nomads" || strings.Contains(site.URL, "workingnomads.com") {
		return crawlWorkingNomadsAPI(site)
	}

	if strings.Contains(strings.ToLower(site.Name), "workana") || strings.Contains(site.URL, "workana.com") {
		return crawlWorkana(site)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", site.URL, nil)
	if err != nil {
		return err
	}
	uaIndex := len(site.URL) % len(userAgents)
	req.Header.Set("User-Agent", userAgents[uaIndex])
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Erro ao acessar %s: %v", site.URL, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("site retornou status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	selector := site.SelectorLinks
	if selector == "" {
		// Fallback para seletores comuns
		selector = "a[href*='job'], a[href*='/jobs/'], a.job-link, article a"
		log.Printf("Seletor vazio, usando padrão: %s", selector)
	}

	log.Printf("Usando seletor de links: %s", selector)

	elemCount := doc.Find(selector).Length()
	log.Printf("Seletor encontrou %d elementos na página", elemCount)

	isLinkedIn := strings.Contains(strings.ToLower(site.Name), "linkedin") ||
		strings.Contains(strings.ToLower(site.URL), "linkedin.com")

	now := time.Now()
	
	if isLinkedIn {
		// Abordagem baseada em cards: extrai título, empresa e link de cada card
		doc.Find(selector).Each(func(i int, card *goquery.Selection) {
			// Extrai link da vaga
			linkHref, _ := card.Find("a[href*='/jobs/view/']").Attr("href")
			if linkHref == "" {
				return
			}
			link := normalizeURL(linkHref, site.URL)

			// Extrai título
			title := strings.TrimSpace(card.Find("h3.base-search-card__title").First().Text())
			if title == "" {
				title = strings.TrimSpace(card.Find(".sr-only").First().Text())
			}
			if title == "" {
				return
			}

			// Extrai empresa
			company := ""
			if site.SelectorCompany != "" {
				company = strings.TrimSpace(card.Find(site.SelectorCompany).First().Text())
			}

			log.Printf("Link encontrado: %s", title)
			if company != "" {
				log.Printf("Empresa: %s", company)
			}

			// Filtra links que não são vagas
			titleLower := strings.ToLower(title)
			if strings.Contains(titleLower, "back to") ||
				strings.Contains(titleLower, "view all") ||
				strings.Contains(titleLower, "login") ||
				strings.Contains(titleLower, "sign up") ||
				strings.Contains(titleLower, "post a job") {
				return
			}

			// Verifica se a vaga já existe
			var existingJob models.Job
			if err := db.DB.Where("url = ?", link).First(&existingJob).Error; err == nil {
				if existingJob.Description == "" {
					log.Printf("Vaga já existe mas sem descrição, buscando detalhes...")
					fetchJobDetails(&existingJob, site, client)
					db.DB.Save(&existingJob)
				}
				return
			}

			job := models.Job{
				OrganizationID: site.OrganizationID,
				SiteID:         &site.ID,
				URL:            link,
				Title:          title,
				Company:        company,
				CollectedAt:    now,
				Status:         models.JobStatusNew,
			}

			if site.DelaySeconds > 0 {
				time.Sleep(time.Duration(site.DelaySeconds) * time.Second)
			}
			fetchJobDetails(&job, site, client)

			if job.Company == "" && company != "" {
				job.Company = company
			}

			if shouldIgnoreJob(&job) {
				log.Printf("Vaga ignorada por filtro: %s", title)
				return
			}

			if err := db.DB.Create(&job).Error; err != nil {
				log.Printf("Erro ao salvar vaga: %v", err)
			} else {
				jobsFound++
				log.Printf("Vaga salva: %s", title)
			}
		})
	} else {
		// Abordagem original baseada em links
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if !exists || href == "" {
				return
			}

			title := strings.TrimSpace(s.Text())
			if title == "" {
				if t, ok := s.Attr("title"); ok && t != "" {
					title = strings.TrimSpace(t)
				} else if t, ok := s.Attr("aria-label"); ok && t != "" {
					title = strings.TrimSpace(t)
				} else {
					title = strings.TrimSpace(s.Find("h2, h3, .title, .job-title").First().Text())
				}
			}

			if title == "" {
				return
			}

			company := ""
			if site.SelectorCompany != "" {
				parent := s.Parent()
				company = strings.TrimSpace(parent.Find(site.SelectorCompany).First().Text())
			}

			link := normalizeURL(href, site.URL)

			log.Printf("Link encontrado: %s - %s", title, link)
			if company != "" {
				log.Printf("Empresa encontrada na listagem: %s", company)
			}

			titleLower := strings.ToLower(title)
			if strings.Contains(titleLower, "back to") ||
				strings.Contains(titleLower, "view all") ||
				strings.Contains(titleLower, "login") ||
				strings.Contains(titleLower, "sign up") ||
				strings.Contains(titleLower, "post a job") {
				return
			}

			var existingJob models.Job
			if err := db.DB.Where("url = ?", link).First(&existingJob).Error; err == nil {
				if existingJob.Description == "" {
					log.Printf("Vaga já existe mas sem descrição, buscando detalhes...")
					fetchJobDetails(&existingJob, site, client)
					db.DB.Save(&existingJob)
				}
				return
			}

			job := models.Job{
				OrganizationID: site.OrganizationID,
				SiteID:         &site.ID,
				URL:            link,
				Title:          title,
				Company:        company,
				CollectedAt:    now,
				Status:         models.JobStatusNew,
			}

			if site.DelaySeconds > 0 {
				time.Sleep(time.Duration(site.DelaySeconds) * time.Second)
			}
			fetchJobDetails(&job, site, client)

			if job.Company == "" && company != "" {
				job.Company = company
			}

			if shouldIgnoreJob(&job) {
				log.Printf("Vaga ignorada por filtro (descrição): %s", title)
				return
			}

			if err := db.DB.Create(&job).Error; err != nil {
				log.Printf("Erro ao salvar vaga: %v", err)
			} else {
				jobsFound++
				log.Printf("Vaga salva: %s", title)
			}
		})
	}

	log.Printf("Vagas encontradas: %d", jobsFound)
	db.DB.Model(&site).Update("last_crawl", time.Now())
	db.Log(site.OrganizationID, "crawler", "crawl_completed", fmt.Sprintf(`{"site": "%s", "url": "%s", "jobs": %d}`, site.Name, site.URL, jobsFound))

	return nil
}

// normalizeURL normaliza URLs relativas e absolutas
func normalizeURL(href, baseURL string) string {
	href = strings.TrimSpace(href)
	
	if strings.HasPrefix(href, "http") {
		// Remove query params do LinkedIn
		if strings.Contains(href, "linkedin.com/jobs/view/") {
			parts := strings.Split(href, "?")
			href = strings.TrimSuffix(parts[0], "/")
			return href
		}
		return href
	}

	// URL relativa
	if strings.HasPrefix(href, "/") {
		// Extrai domínio do baseURL
		parts := strings.Split(baseURL, "/")
		if len(parts) >= 3 {
			return parts[0] + "//" + parts[2] + href
		}
	}

	// Relativa ao diretório atual
	if !strings.HasSuffix(baseURL, "/") {
		baseURL = baseURL + "/"
	}
	return baseURL + href
}

func fetchJobDetails(job *models.Job, site models.Site, client *http.Client) error {
	log.Printf("Buscando detalhes em: %s", job.URL)

	req, err := http.NewRequest("GET", job.URL, nil)
	if err != nil {
		return err
	}
	uaIndex := len(job.URL) % len(userAgents)
	req.Header.Set("User-Agent", userAgents[uaIndex])
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("erro ao buscar detalhes: status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	// Extrai empresa
	company := ""
	companySelectors := []string{site.SelectorCompany}
	if site.SelectorCompany == "" {
		companySelectors = []string{
			".topcard__org-name-link",
			".company-name",
			".job-result-card__subtitle-link",
			"[data-automation-id='job-organic-result-company-name']",
			".company",
		}
	}
	for _, sel := range companySelectors {
		if sel == "" {
			continue
		}
		text := strings.TrimSpace(doc.Find(sel).First().Text())
		if text != "" {
			company = text
			break
		}
	}
	job.Company = company

	// Extrai descrição
	desc := ""
	descSelectors := []string{site.SelectorDescription}
	if site.SelectorDescription == "" {
		descSelectors = []string{
			".show-more-less-html__markup",
			".job-description",
			".description__text",
			"article",
			".job-view-layout",
			"main",
		}
	}
	for _, sel := range descSelectors {
		if sel == "" {
			continue
		}
		text := strings.TrimSpace(doc.Find(sel).First().Text())
		if text != "" {
			desc = text
			break
		}
	}
	job.Description = desc

	return nil
}

func crawlWorkana(site models.Site) error {
	log.Printf("Usando crawler específico para Workana")

	client := &http.Client{Timeout: 30 * time.Second}
	now := time.Now()
	var jobsFound int

	resultsPattern := regexp.MustCompile(`results-initials=(['"])`)

	baseURL := site.URL
	if !strings.Contains(baseURL, "workana.com") {
		baseURL = "https://www.workana.com/jobs"
	}

	for page := 1; page <= 50; page++ {
		var pageURL string
		if strings.Contains(baseURL, "?") {
			pageURL = baseURL + "&page=" + strconv.Itoa(page)
		} else {
			pageURL = baseURL + "?page=" + strconv.Itoa(page)
		}

		log.Printf("Buscando página %d: %s", page, pageURL)

		req, err := http.NewRequest("GET", pageURL, nil)
		if err != nil {
			return err
		}
		uaIndex := len(pageURL) % len(userAgents)
		req.Header.Set("User-Agent", userAgents[uaIndex])
		req.Header.Set("Accept", "text/html,application/xhtml+xml")
		req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7")

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Erro ao acessar página %d: %v", page, err)
			break
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		htmlContent := string(body)

		if resp.StatusCode != 200 {
			log.Printf("Página %d retornou status %d", page, resp.StatusCode)
			break
		}

		loc := resultsPattern.FindStringSubmatchIndex(htmlContent)
		if loc == nil {
			log.Printf("Nenhum dado encontrado na página %d", page)
			break
		}
		// value starts right after the opening quote
		quoteByte := htmlContent[loc[2]]
		valStart := loc[2] + 1
		valEnd := strings.IndexByte(htmlContent[valStart:], byte(quoteByte))
		if valEnd == -1 {
			log.Printf("Aspas de fechamento não encontradas na página %d", page)
			break
		}
		rawJSON := htmlContent[valStart : valStart+valEnd]

		decoded := html.UnescapeString(rawJSON)
		var pageData map[string]interface{}
		if err := json.Unmarshal([]byte(decoded), &pageData); err != nil {
			log.Printf("Erro ao decodificar JSON na página %d: %v", page, err)
			break
		}

		resultsRaw, ok := pageData["results"].([]interface{})
		if !ok || len(resultsRaw) == 0 {
			log.Printf("Nenhum resultado na página %d", page)
			break
		}

		log.Printf("Encontrados %d resultados na página %d", len(resultsRaw), page)

		for _, item := range resultsRaw {
			result, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			slug, _ := result["slug"].(string)
			if slug == "" {
				continue
			}

			titleRaw, _ := result["title"].(string)
			title := stripHTML(titleRaw)
			title = strings.TrimSpace(title)
			if title == "" {
				title = slug
			}

			authorName, _ := result["authorName"].(string)
			descRaw, _ := result["description"].(string)
			description := stripHTML(descRaw)

			jobURL := "https://www.workana.com/job/" + slug

			var existingJob models.Job
			if err := db.DB.Where("url = ?", jobURL).First(&existingJob).Error; err == nil {
				continue
			}

			job := models.Job{
				OrganizationID: site.OrganizationID,
				SiteID:         &site.ID,
				URL:            jobURL,
				Title:          title,
				Company:        authorName,
				Description:    description,
				CollectedAt:    now,
				Status:         models.JobStatusNew,
			}

			if shouldIgnoreJob(&job) {
				log.Printf("Vaga ignorada por filtro: %s", title)
				continue
			}

			if err := db.DB.Create(&job).Error; err != nil {
				log.Printf("Erro ao salvar vaga: %v", err)
			} else {
				jobsFound++
				log.Printf("Vaga salva: %s", title)
			}
		}

		if len(resultsRaw) == 0 {
			log.Printf("Nenhum resultado na página %d, encerrando paginação", page)
			break
		}

		if site.DelaySeconds > 0 {
			time.Sleep(time.Duration(site.DelaySeconds) * time.Second)
		}
	}

	log.Printf("Total de vagas encontradas na Workana: %d", jobsFound)
	db.DB.Model(&site).Update("last_crawl", time.Now())
	db.Log(site.OrganizationID, "crawler", "crawl_completed", fmt.Sprintf(`{"site": "%s", "jobs": %d}`, site.Name, jobsFound))

	return nil
}

func crawlRemoteOKAPI(site models.Site) error {
	log.Printf("Usando API para RemoteOK: https://remoteok.com/api")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", "https://remoteok.com/api", nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgents[0])
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Erro ao acessar API: %v", err)
		return err
	}
	defer resp.Body.Close()

	var jobs []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&jobs); err != nil {
		log.Printf("Erro ao decodificar JSON: %v", err)
		return err
	}

	if len(jobs) == 0 {
		log.Println("Nenhuma vaga encontrada na API")
		return nil
	}

	now := time.Now()
	var jobsFound int

	for _, item := range jobs {
		id, ok := item["id"].(string)
		if !ok || id == "" {
			continue
		}

		company, _ := item["company"].(string)
		title, _ := item["position"].(string)
		jobURL, _ := item["url"].(string)

		if jobURL == "" {
			jobURL = "https://remoteok.com/remote-jobs/" + id
		}

		var job models.Job
		if err := db.DB.Where("url = ?", jobURL).First(&job).Error; err == nil {
			continue
		}

		description := fmt.Sprintf("%s - %s", company, title)
		if val, ok := item["description"].(string); ok {
			description = val
		}

		job = models.Job{
			OrganizationID: site.OrganizationID,
			SiteID:        &site.ID,
			URL:           jobURL,
			Title:         title,
			Company:       company,
			Description:   description,
			CollectedAt:   now,
			Status:        models.JobStatusNew,
		}

		if err := db.DB.Create(&job).Error; err != nil {
			log.Printf("Erro ao salvar vaga: %v", err)
		} else {
			jobsFound++
			log.Printf("Vaga encontrada: %s @ %s", title, company)
		}
	}

	log.Printf("Total de vagas salvas: %d", jobsFound)
	db.DB.Model(&site).Update("last_crawl", now)
	db.Log(site.OrganizationID, "crawler", "crawl_completed", fmt.Sprintf(`{"site": "%s", "jobs": %d}`, site.Name, jobsFound))

	return nil
}

func crawlWorkingNomadsAPI(site models.Site) error {
	log.Printf("Usando API para Working Nomads: https://www.workingnomads.com/api/exposed_jobs/")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", "https://www.workingnomads.com/api/exposed_jobs/", nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgents[0])
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Erro ao acessar API: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("site retornou status %d", resp.StatusCode)
	}

	var jobs []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&jobs); err != nil {
		log.Printf("Erro ao decodificar JSON: %v", err)
		return err
	}

	if len(jobs) == 0 {
		log.Println("Nenhuma vaga encontrada na API")
		return nil
	}

	now := time.Now()
	var jobsFound int

	for _, item := range jobs {
		jobURL, ok := item["url"].(string)
		if !ok || jobURL == "" {
			continue
		}

		title, _ := item["title"].(string)
		description, _ := item["description"].(string)

		if title == "" {
			continue
		}

		var job models.Job
		if err := db.DB.Where("url = ?", jobURL).First(&job).Error; err == nil {
			continue
		}

		job = models.Job{
			OrganizationID: site.OrganizationID,
			SiteID:        &site.ID,
			URL:           jobURL,
			Title:         title,
			Description:   stripHTML(description),
			CollectedAt:   now,
			Status:        models.JobStatusNew,
		}

		if err := db.DB.Create(&job).Error; err != nil {
			log.Printf("Erro ao salvar vaga: %v", err)
		} else {
			jobsFound++
			log.Printf("Vaga encontrada: %s", title)
		}
	}

	log.Printf("Total de vagas salvas: %d", jobsFound)
	db.DB.Model(&site).Update("last_crawl", now)
	db.Log(site.OrganizationID, "crawler", "crawl_completed", fmt.Sprintf(`{"site": "%s", "jobs": %d}`, site.Name, jobsFound))

	return nil
}

func stripHTML(html string) string {
	re := regexp.MustCompile("<[^>]*>")
	return strings.TrimSpace(re.ReplaceAllString(html, ""))
}

func shouldIgnoreJob(job *models.Job) bool {
	titleLower := strings.ToLower(job.Title)
	descLower := strings.ToLower(job.Description)

	// Filtros de senior
	if strings.Contains(titleLower, "senior only") ||
		strings.Contains(titleLower, "senior-level") ||
		strings.Contains(titleLower, "sr.") ||
		strings.Contains(titleLower, "sr ") ||
		strings.Contains(descLower, "senior only") ||
		strings.Contains(descLower, "senior-level") ||
		strings.Contains(descLower, "sr.") {
		return true
	}

	// Filtros de remote only
	if strings.Contains(titleLower, "remote only") ||
		strings.Contains(titleLower, "100% remote") ||
		strings.Contains(descLower, "remote only") ||
		strings.Contains(descLower, "100% remote") {
		return true
	}

	return false
}
