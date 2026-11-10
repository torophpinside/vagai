package crawler

import (
	"encoding/json"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const maxLocationLen = 255

// locationSelectors: fallback quando a página não traz JSON-LD (JobPosting).
var locationSelectors = []string{
	"span.topcard__flavor.topcard__flavor--bullet", // LinkedIn (página pública)
	".job-details-jobs-unified-top-card__bullet",
	"[itemprop='jobLocation']",
	".job-location",
	".location",
}

// extractLocation lê cidade/UF e modelo de trabalho da página da vaga.
// Prioridade: JSON-LD JobPosting (estruturado) e depois seletores CSS.
func extractLocation(doc *goquery.Document) (location, workModel string) {
	doc.Find(`script[type="application/ld+json"]`).EachWithBreak(func(_ int, s *goquery.Selection) bool {
		loc, model := locationFromJSONLD(s.Text())
		if loc != "" || model != "" {
			location, workModel = loc, model
			return false
		}
		return true
	})

	if location == "" {
		for _, sel := range locationSelectors {
			text := cleanLocation(doc.Find(sel).First().Text())
			if text != "" {
				location = text
				break
			}
		}
	}
	if workModel == "" {
		workModel = detectWorkModel(location)
	}
	return location, workModel
}

func cleanLocation(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > maxLocationLen {
		s = s[:maxLocationLen]
	}
	return s
}

// detectWorkModel infere o modelo a partir de um texto curto (ex.: "Curitiba, PR (Híbrido)").
func detectWorkModel(s string) string {
	l := strings.ToLower(s)
	switch {
	case strings.Contains(l, "híbrido"), strings.Contains(l, "hibrido"), strings.Contains(l, "hybrid"):
		return "hybrid"
	case strings.Contains(l, "presencial"), strings.Contains(l, "on-site"), strings.Contains(l, "onsite"):
		return "presencial"
	case strings.Contains(l, "remoto"), strings.Contains(l, "remote"):
		return "remote"
	}
	return ""
}

func locationFromJSONLD(raw string) (location, workModel string) {
	var data interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &data); err != nil {
		return "", ""
	}
	for _, posting := range findJobPostings(data) {
		if t, _ := posting["jobLocationType"].(string); strings.EqualFold(t, "TELECOMMUTE") {
			workModel = "remote"
		}
		var parts []string
		switch jl := posting["jobLocation"].(type) {
		case map[string]interface{}:
			parts = append(parts, addressText(jl))
		case []interface{}:
			for _, item := range jl {
				if m, ok := item.(map[string]interface{}); ok {
					parts = append(parts, addressText(m))
				}
			}
		}
		var nonEmpty []string
		for _, p := range parts {
			if p != "" {
				nonEmpty = append(nonEmpty, p)
			}
		}
		return cleanLocation(strings.Join(nonEmpty, " | ")), workModel
	}
	return "", ""
}

func findJobPostings(v interface{}) []map[string]interface{} {
	var out []map[string]interface{}
	switch t := v.(type) {
	case []interface{}:
		for _, item := range t {
			out = append(out, findJobPostings(item)...)
		}
	case map[string]interface{}:
		if typ, _ := t["@type"].(string); typ == "JobPosting" {
			out = append(out, t)
		}
		if graph, ok := t["@graph"]; ok {
			out = append(out, findJobPostings(graph)...)
		}
	}
	return out
}

func addressText(place map[string]interface{}) string {
	addr, ok := place["address"].(map[string]interface{})
	if !ok {
		if s, ok := place["address"].(string); ok {
			return s
		}
		return ""
	}
	var parts []string
	for _, k := range []string{"addressLocality", "addressRegion"} {
		if s, _ := addr[k].(string); strings.TrimSpace(s) != "" {
			parts = append(parts, strings.TrimSpace(s))
		}
	}
	switch c := addr["addressCountry"].(type) {
	case string:
		if strings.TrimSpace(c) != "" {
			parts = append(parts, strings.TrimSpace(c))
		}
	case map[string]interface{}:
		if s, _ := c["name"].(string); strings.TrimSpace(s) != "" {
			parts = append(parts, strings.TrimSpace(s))
		}
	}
	return strings.Join(parts, ", ")
}
