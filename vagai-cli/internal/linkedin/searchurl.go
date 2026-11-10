// Package linkedin builds LinkedIn job search URLs from structured filters.
package linkedin

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"
)

// BaseURL is the LinkedIn job search endpoint.
const BaseURL = "https://www.linkedin.com/jobs/search/"

const (
	maxTextLen       = 100
	maxCompanies     = 10
	maxCompanyLength = 50
)

// DatePosted is the recency filter.
type DatePosted string

// ExperienceLevel is the seniority filter.
type ExperienceLevel string

// RemoteFilter is the work arrangement filter.
type RemoteFilter string

const (
	Past24Hours DatePosted = "past_24_hours"
	PastWeek    DatePosted = "past_week"
	PastMonth   DatePosted = "past_month"

	EntryLevel ExperienceLevel = "entry_level"
	Associate  ExperienceLevel = "associate"
	MidSenior  ExperienceLevel = "mid_senior"
	Manager    ExperienceLevel = "manager"
	Director   ExperienceLevel = "director"

	OnSite RemoteFilter = "on_site"
	Remote RemoteFilter = "remote"
	Hybrid RemoteFilter = "hybrid"
)

// f_TPR takes seconds since posting: "r" + seconds.
var dateCodes = map[DatePosted]string{
	Past24Hours: "r86400",
	PastWeek:    "r604800",
	PastMonth:   "r2592000",
}

// LinkedIn f_E: 1 internship, 2 entry, 3 associate, 4 mid-senior, 5 director, 6 executive.
// LinkedIn has no separate "manager" level; it maps to mid-senior.
var experienceCodes = map[ExperienceLevel]string{
	EntryLevel: "2",
	Associate:  "3",
	MidSenior:  "4",
	Manager:    "4",
	Director:   "5",
}

// LinkedIn f_WT: 1 on-site, 2 remote, 3 hybrid.
var remoteCodes = map[RemoteFilter]string{
	OnSite: "1",
	Remote: "2",
	Hybrid: "3",
}

// Query describes a LinkedIn job search. All fields are optional.
type Query struct {
	Keywords        string
	Location        string
	DatePosted      DatePosted
	ExperienceLevel ExperienceLevel
	RemoteFilter    RemoteFilter
	// CompanyFilters are company names appended to the keywords.
	CompanyFilters []string
}

// Validate checks the query against the data-model rules.
func (q Query) Validate() error {
	if err := checkText("keywords", q.Keywords, maxTextLen); err != nil {
		return err
	}
	if err := checkText("location", q.Location, maxTextLen); err != nil {
		return err
	}
	if q.DatePosted != "" {
		if _, ok := dateCodes[q.DatePosted]; !ok {
			return fmt.Errorf("invalid date_posted %q", q.DatePosted)
		}
	}
	if q.ExperienceLevel != "" {
		if _, ok := experienceCodes[q.ExperienceLevel]; !ok {
			return fmt.Errorf("invalid experience_level %q", q.ExperienceLevel)
		}
	}
	if q.RemoteFilter != "" {
		if _, ok := remoteCodes[q.RemoteFilter]; !ok {
			return fmt.Errorf("invalid remote_filter %q", q.RemoteFilter)
		}
	}
	if len(q.CompanyFilters) > maxCompanies {
		return fmt.Errorf("company_filters: at most %d items", maxCompanies)
	}
	for _, c := range q.CompanyFilters {
		if strings.TrimSpace(c) == "" {
			return errors.New("company_filters: empty item")
		}
		if err := checkText("company_filters item", c, maxCompanyLength); err != nil {
			return err
		}
	}
	return nil
}

func checkText(name, v string, max int) error {
	if v == "" {
		return nil
	}
	if strings.TrimSpace(v) == "" {
		return fmt.Errorf("%s: must not be blank", name)
	}
	if utf8.RuneCountInString(strings.TrimSpace(v)) > max {
		return fmt.Errorf("%s: at most %d characters", name, max)
	}
	return nil
}

// BuildSearchURL returns a LinkedIn job search URL for q.
// Values are percent-encoded per RFC 3986 (spaces as %20).
func BuildSearchURL(q Query) (string, error) {
	if err := q.Validate(); err != nil {
		return "", err
	}

	var parts []string
	add := func(k, v string) {
		parts = append(parts, k+"="+encode(v))
	}

	keywords := strings.TrimSpace(q.Keywords)
	for _, c := range q.CompanyFilters {
		if keywords != "" {
			keywords += " "
		}
		keywords += strings.TrimSpace(c)
	}
	if keywords != "" {
		add("keywords", keywords)
	}
	if loc := strings.TrimSpace(q.Location); loc != "" {
		add("location", loc)
	}
	if q.DatePosted != "" {
		add("f_TPR", dateCodes[q.DatePosted])
	}
	if q.ExperienceLevel != "" {
		add("f_E", experienceCodes[q.ExperienceLevel])
	}
	if q.RemoteFilter != "" {
		add("f_WT", remoteCodes[q.RemoteFilter])
	}

	if len(parts) == 0 {
		return BaseURL, nil
	}
	return BaseURL + "?" + strings.Join(parts, "&"), nil
}

// ValidateURL reports whether raw is a well-formed LinkedIn job search URL.
func ValidateURL(raw string) error {
	if !strings.HasPrefix(raw, BaseURL) {
		return fmt.Errorf("url must start with %s", BaseURL)
	}
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if _, err := url.ParseQuery(u.RawQuery); err != nil {
		return err
	}
	return nil
}

func encode(s string) string {
	return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
}
