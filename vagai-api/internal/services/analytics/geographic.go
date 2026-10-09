package analytics

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/anomalyco/vagai-api/internal/models"
	"gorm.io/gorm"
)

// Mensagens de estado vazio do geographic-insights (T069) — espelham
// EmptyResumeMessage/EmptyResumeGuidance (resume.go).
const (
	EmptyGeographicMessage  = "Nenhuma informação geográfica disponível"
	EmptyGeographicGuidance = "Cadastre usuários na organização e colete vagas (crawler) para começar a ver a distribuição geográfica e a razão de trabalho remoto."
)

// geoRuneFold mapeia os acentos comuns do português para a letra base
// (T066 — normalização de nomes de cidade no lado Go).
var geoRuneFold = map[rune]rune{
	'á': 'a', 'à': 'a', 'â': 'a', 'ã': 'a',
	'é': 'e', 'ê': 'e',
	'í': 'i',
	'ó': 'o', 'ô': 'o', 'õ': 'o',
	'ú': 'u',
	'ç': 'c',
}

// geoCityToUF mapeia cidades brasileiras (minúsculas e sem acentos) para a
// UF. T066: jobs não têm coluna de cidade — a localização é inferida por
// LIKE sobre título + descrição normalizados.
var geoCityToUF = map[string]string{
	// Capitais (27 + DF)
	"sao paulo":      "SP",
	"rio de janeiro": "RJ",
	"brasilia":       "DF",
	"salvador":       "BA",
	"fortaleza":      "CE",
	"belo horizonte": "MG",
	"manaus":         "AM",
	"curitiba":       "PR",
	"recife":         "PE",
	"goiania":        "GO",
	"belem":          "PA",
	"porto alegre":   "RS",
	"maceio":         "AL",
	"campo grande":   "MS",
	"joao pessoa":    "PB",
	"natal":          "RN",
	"teresina":       "PI",
	"cuiaba":         "MT",
	"aracaju":        "SE",
	"florianopolis":  "SC",
	"vitoria":        "ES",
	"sao luis":       "MA",
	"palmas":         "TO",
	"macapa":         "AP",
	"boa vista":      "RR",
	"porto velho":    "RO",
	"rio branco":     "AC",
	// Demais cidades de grande porte
	"guarulhos":             "SP",
	"campinas":              "SP",
	"sao bernardo do campo": "SP",
	"sao jose dos campos":   "SP",
	"santos":                "SP",
	"sao jose do rio preto": "SP",
	"osasco":                "SP",
	"ribeirao preto":        "SP",
	"sorocaba":              "SP",
	"bauru":                 "SP",
	"juiz de fora":          "MG",
	"uberlandia":            "MG",
	"montes claros":         "MG",
	"niteroi":               "RJ",
	"nova iguacu":           "RJ",
	"duque de caxias":       "RJ",
	"volta redonda":         "RJ",
	"sao goncalo":           "RJ",
	"londrina":              "PR",
	"maringa":               "PR",
	"ponta grossa":          "PR",
	"joinville":             "SC",
	"blumenau":              "SC",
	"pelotas":               "RS",
	"canoas":                "RS",
	"feira de santana":      "BA",
	"vila velha":            "ES",
	"petrolina":             "PE",
	"campina grande":        "PB",
}

// Geographic consolida os insights geográficos e demográficos (US10 —
// T065–T067): usuários por cidade/fuso horário, vagas por cidade/UF
// inferidas do texto e a razão de vagas amigáveis a trabalho remoto.
//
// Retorna nil quando a organização (ou a visão global) não possui usuários
// com membership nem vagas (estado vazio, T069).
func (s *Service) Geographic(ctx context.Context, userID uint) (*models.GeographicInsights, error) {
	db, ctx, cancel := s.ctxDB(ctx)
	defer cancel()

	key := fmt.Sprintf("org:%d:geographic", userID)
	return cachedValue(ctx, s, key, func(ctx context.Context) (*models.GeographicInsights, error) {
		return loadGeographic(db, userID)
	})
}

func loadGeographic(db *gorm.DB, userID uint) (*models.GeographicInsights, error) {
	// T069 — estado vazio: zero usuários (com membership) E zero vagas.
	users, err := geoCountUsers(db, userID)
	if err != nil {
		return nil, err
	}
	jobs, err := geoCountJobs(db, userID)
	if err != nil {
		return nil, err
	}
	if users == 0 && jobs == 0 {
		return nil, nil
	}

	// T065 — usuários por cidade (top 50) e por fuso horário (top 30).
	cityRows, err := geoUserAggregate(db, userID, "u.city", 50)
	if err != nil {
		return nil, err
	}
	tzRows, err := geoUserAggregate(db, userID, "u.timezone", 30)
	if err != nil {
		return nil, err
	}
	usersByCity := make([]models.CityCount, 0, len(cityRows))
	for _, r := range cityRows {
		usersByCity = append(usersByCity, models.CityCount{City: r.Name, Count: r.Cnt})
	}
	usersByTimezone := make([]models.TimezoneCount, 0, len(tzRows))
	for _, r := range tzRows {
		usersByTimezone = append(usersByTimezone, models.TimezoneCount{Timezone: r.Name, Count: r.Cnt})
	}

	// T066 — vagas por cidade/UF inferidas do texto.
	jobsByCity, jobsByState, err := geoJobsByLocation(db, userID)
	if err != nil {
		return nil, err
	}

	// T067 — razão de vagas amigáveis a trabalho remoto.
	remoteRatio, remoteToOnSite, err := geoRemoteRatios(db, userID)
	if err != nil {
		return nil, err
	}

	return &models.GeographicInsights{
		UsersByCity:         usersByCity,
		UsersByTimezone:     usersByTimezone,
		JobsByCity:          jobsByCity,
		JobsByState:         jobsByState,
		RemoteFriendlyRatio: remoteRatio,
		RemoteToOnSiteRatio: remoteToOnSite,
	}, nil
}

// geoNameCount é a linha genérica das agregações de usuários (name + cnt).
type geoNameCount struct {
	Name string
	Cnt  int64
}

// geoCountUsers conta usuários vivos (escopados pelo usuário quando
// userID > 0; userID == 0 agrega todos os usuários da plataforma).
func geoCountUsers(db *gorm.DB, userID uint) (int64, error) {
	q := `SELECT COUNT(*) FROM users u WHERE u.deleted_at IS NULL`
	args := []any{}
	if userID > 0 {
		q += " AND u.id = ?"
		args = append(args, userID)
	}

	var total int64
	if err := db.Raw(q, args...).Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// geoCountJobs conta as vagas da organização (global quando userID == 0).
func geoCountJobs(db *gorm.DB, userID uint) (int64, error) {
	q := db.Table("jobs")
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// geoUserAggregate conta usuários distintos agrupados por column
// ("u.city" ou "u.timezone"), ignorando valores vazios. Com userID == 0 o
// JOIN sem filtro de organização agrega no global sem duplicar usuários
// (COUNT(DISTINCT u.id)).
func geoUserAggregate(db *gorm.DB, userID uint, column string, limit int) ([]geoNameCount, error) {
	q := fmt.Sprintf(`SELECT %s AS name, COUNT(*) AS cnt
		FROM users u`, column)
	args := []any{}
	if userID > 0 {
		q += " WHERE u.id = ? AND u.deleted_at IS NULL"
		args = append(args, userID)
	} else {
		q += " WHERE u.deleted_at IS NULL"
	}
	q += fmt.Sprintf(` AND %s <> ''
		GROUP BY %s ORDER BY cnt DESC LIMIT %d`, column, column, limit)

	var rows []geoNameCount
	if err := db.Raw(q, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// geoCityRef é o par nome normalizado → UF de uma cidade do mapa.
type geoCityRef struct {
	Name string
	UF   string
}

// geoCityRefs devolve as cidades do geoCityToUF em ordem estável, com nomes
// normalizados (geoNormalize é idempotente: minúsculas + sem acentos) e sem
// duplicatas.
func geoCityRefs() []geoCityRef {
	byName := make(map[string]string, len(geoCityToUF))
	for city, uf := range geoCityToUF {
		byName[geoNormalize(city)] = uf
	}
	refs := make([]geoCityRef, 0, len(byName))
	for name, uf := range byName {
		refs = append(refs, geoCityRef{Name: name, UF: uf})
	}
	sort.Slice(refs, func(i, j int) bool { return refs[i].Name < refs[j].Name })
	return refs
}

// geoJobsByLocation varre as vagas uma única vez em Go — strings.Contains
// (SIMD) sobre o texto normalizado uma vez por linha (T066) — e deriva o
// total por UF a partir das cidades casadas. LIKE '%cidade%' no MySQL é
// O(linha×cidade×bytes) e estoura o budget de 5s em descrições longas.
func geoJobsByLocation(db *gorm.DB, userID uint) ([]models.CityCount, []models.StateCount, error) {
	type jobText struct {
		Title       string
		Description string
	}
	q := db.Model(&models.Job{}).Select("title, description")
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	var jobs []jobText
	if err := q.Find(&jobs).Error; err != nil {
		return nil, nil, err
	}

	refs := geoCityRefs()
	counts := make([]int64, len(refs))
	for _, job := range jobs {
		text := geoNormalize(job.Title + " " + job.Description)
		for i, ref := range refs {
			if strings.Contains(text, ref.Name) {
				counts[i]++
			}
		}
	}

	byUF := make(map[string]int64, len(refs))
	byCity := make([]models.CityCount, 0, len(refs))
	for i, ref := range refs {
		if counts[i] <= 0 {
			continue
		}
		byCity = append(byCity, models.CityCount{City: ref.Name, Count: counts[i]})
		byUF[ref.UF] += counts[i]
	}
	sort.Slice(byCity, func(a, b int) bool {
		if byCity[a].Count != byCity[b].Count {
			return byCity[a].Count > byCity[b].Count
		}
		return byCity[a].City < byCity[b].City
	})
	if len(byCity) > 50 {
		byCity = byCity[:50]
	}

	byState := make([]models.StateCount, 0, len(byUF))
	for uf, n := range byUF {
		byState = append(byState, models.StateCount{State: uf, Count: n})
	}
	sort.Slice(byState, func(a, b int) bool {
		if byState[a].Count != byState[b].Count {
			return byState[a].Count > byState[b].Count
		}
		return byState[a].State < byState[b].State
	})
	return byCity, byState, nil
}

// geoRemoteRow guarda os totais brutos da query de trabalho remoto.
type geoRemoteRow struct {
	Total  float64
	Remote float64
}

// geoRemoteRatios calcula a razão de vagas remotas em uma única query
// (T067): RemoteFriendlyRatio = remoto/total arredondado em 2 casas e
// RemoteToOnSiteRatio = "r:o" irreduzível pela GCD.
func geoRemoteRatios(db *gorm.DB, userID uint) (float64, string, error) {
	q := `SELECT COUNT(*) AS total,
		COALESCE(SUM(LOWER(CONCAT(COALESCE(title,''),' ',COALESCE(description,'')))
			REGEXP 'remote|remoto|home office|teletrabalho|hibrido remoto'),0) AS remote
		FROM jobs`
	args := []any{}
	if userID > 0 {
		q += " WHERE user_id = ?"
		args = append(args, userID)
	}

	var row geoRemoteRow
	if err := db.Raw(q, args...).Scan(&row).Error; err != nil {
		return 0, "", err
	}
	total := int64(row.Total)
	remote := int64(row.Remote)
	onsite := total - remote
	if onsite < 0 {
		onsite = 0
	}
	return round2(ratio(remote, total)), geoRatioString(remote, onsite), nil
}

// geoRatioString formata "remoto:presencial" reduzido pela GCD
// (78:22 → "39:11", 100:0 → "1:0", 0:0 → "0:0").
func geoRatioString(remote, onsite int64) string {
	g := geoGCD(remote, onsite)
	if g <= 1 {
		return fmt.Sprintf("%d:%d", remote, onsite)
	}
	return fmt.Sprintf("%d:%d", remote/g, onsite/g)
}

// geoGCD devolve o máximo divisor comum (algoritmo de Euclides; GCD(0,0)=0).
func geoGCD(a, b int64) int64 {
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// geoNormalize deixa a string em minúsculas e remove os acentos de
// geoRuneFold — minúsculas + sem acentos casam com as chaves do mapa e com o
// haystack normalizado no SQL (idempotente).
func geoNormalize(s string) string {
	return strings.Map(func(r rune) rune {
		if v, ok := geoRuneFold[r]; ok {
			return v
		}
		return r
	}, strings.ToLower(s))
}
