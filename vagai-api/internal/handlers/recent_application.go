package handlers

import (
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/anomalyco/vagai-api/internal/models"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"gorm.io/gorm"
)

// Janela em que uma candidatura conta como "recente" para o aviso de mesma empresa.
const recentApplicationDays = 30

var (
	companyNonAlnum = regexp.MustCompile(`[^a-z0-9]+`)
	companySuffix   = regexp.MustCompile(`\s(ltda|inc|llc|ltd|corp|sa|me|eireli)$`)
)

// normalizeCompany: trim, lowercase, sem acento, sem sufixo societário final,
// pontuação e espaços colapsados.
func normalizeCompany(name string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	s, _, err := transform.String(t, strings.ToLower(strings.TrimSpace(name)))
	if err != nil {
		s = strings.ToLower(strings.TrimSpace(name))
	}
	s = strings.ReplaceAll(s, ".", "") // "s.a." -> "sa"
	s = strings.TrimSpace(companyNonAlnum.ReplaceAllString(s, " "))
	for {
		next := companySuffix.ReplaceAllString(s, "")
		if next == s {
			break
		}
		s = next
	}
	return s
}

type recentAppRow struct {
	UserID    uint
	Company   string
	Title     string
	AppliedAt time.Time
}

// attachRecentApplications preenche RecentApplication nos matches não aplicados
// cuja empresa (normalizada) tem candidatura do mesmo usuário dentro da janela.
func attachRecentApplications(db *gorm.DB, matches []models.Match, now time.Time) {
	userIDs := map[uint]struct{}{}
	for _, m := range matches {
		if !m.Applied {
			userIDs[m.UserID] = struct{}{}
		}
	}
	if len(userIDs) == 0 {
		return
	}
	ids := make([]uint, 0, len(userIDs))
	for id := range userIDs {
		ids = append(ids, id)
	}

	var rows []recentAppRow
	db.Table("matches").
		Select("matches.user_id AS user_id, jobs.company AS company, jobs.title AS title, matches.applied_at AS applied_at").
		Joins("JOIN jobs ON jobs.id = matches.job_id").
		Where("matches.user_id IN ?", ids).
		Where("matches.applied = ?", true).
		Where("matches.applied_at IS NOT NULL AND matches.applied_at >= ?", now.AddDate(0, 0, -recentApplicationDays)).
		Where("jobs.status NOT IN ?", []string{"ignored", "unmatched"}).
		Scan(&rows)

	type key struct {
		user    uint
		company string
	}
	latest := map[key]recentAppRow{}
	for _, r := range rows {
		c := normalizeCompany(r.Company)
		if c == "" {
			continue
		}
		k := key{r.UserID, c}
		if cur, ok := latest[k]; !ok || r.AppliedAt.After(cur.AppliedAt) {
			latest[k] = r
		}
	}

	for i := range matches {
		m := &matches[i]
		if m.Applied {
			continue
		}
		c := normalizeCompany(m.Job.Company)
		if c == "" {
			continue
		}
		if r, ok := latest[key{m.UserID, c}]; ok {
			m.RecentApplication = &models.RecentApplication{
				JobTitle:  r.Title,
				Company:   r.Company,
				AppliedAt: r.AppliedAt,
			}
		}
	}
}
