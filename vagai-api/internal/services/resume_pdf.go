package services

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/jung-kurt/gofpdf/v2"
)

type PersonalInfo struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Location string `json:"location"`
	Linkedin string `json:"linkedin"`
	Website  string `json:"website"`
}

type ExperienceEntry struct {
	Company     string `json:"company"`
	Role        string `json:"role"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Description string `json:"description"`
}

type EducationEntry struct {
	Institution string `json:"institution"`
	Degree      string `json:"degree"`
	Field       string `json:"field"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Notes       string `json:"notes"`
}

type ResumeData struct {
	PersonalInfo    PersonalInfo     `json:"personal_info"`
	Summary         string           `json:"summary"`
	Experience      []ExperienceEntry `json:"experience"`
	Education       []EducationEntry  `json:"education"`
	Skills          []string          `json:"skills"`
	Languages       []string          `json:"languages"`
	Certifications  []string          `json:"certifications"`
}

func GenerateResumePDF(data ResumeData) ([]byte, error) {
	experience := append([]ExperienceEntry(nil), data.Experience...)
	sortExperience(experience)

	education := append([]EducationEntry(nil), data.Education...)
	sortEducation(education)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	marginLeft := 15.0
	marginRight := 15.0
	pageW, _ := pdf.GetPageSize()
	pageW = pageW - marginLeft - marginRight
	x := marginLeft

	pdf.SetMargins(marginLeft, 12, marginRight)

	// --- Header: Name ---
	pdf.SetFont("Helvetica", "B", 16)
	pdf.SetXY(x, 12)
	pdf.CellFormat(pageW, 8, sanitizeText(data.PersonalInfo.Name), "", 1, "L", false, 0, "")
	y := pdf.GetY() + 2

	// Contact line
	var contactParts []string
	if data.PersonalInfo.Email != "" {
		contactParts = append(contactParts, data.PersonalInfo.Email)
	}
	if data.PersonalInfo.Phone != "" {
		contactParts = append(contactParts, data.PersonalInfo.Phone)
	}
	if data.PersonalInfo.Location != "" {
		contactParts = append(contactParts, data.PersonalInfo.Location)
	}
	if data.PersonalInfo.Linkedin != "" {
		contactParts = append(contactParts, data.PersonalInfo.Linkedin)
	}
	if data.PersonalInfo.Website != "" {
		contactParts = append(contactParts, data.PersonalInfo.Website)
	}
	if len(contactParts) > 0 {
		pdf.SetFont("Helvetica", "", 10)
		pdf.SetXY(x, y)
		pdf.MultiCell(pageW, 5, sanitizeText(strings.Join(contactParts, " | ")), "", "L", false)
		y = pdf.GetY() + 2
	}

	// Divider
	drawLine(pdf, x, y, x+pageW)
	y += 4

	// --- Summary ---
	if data.Summary != "" {
		y = drawSectionHeader(pdf, x, y, pageW, "RESUMO")
		pdf.SetFont("Helvetica", "", 10)
		pdf.SetXY(x, y)
		pdf.MultiCell(pageW, 5, sanitizeText(data.Summary), "", "L", false)
		y = pdf.GetY() + 4
	}

	// --- Experience ---
	if len(experience) > 0 {
		y = drawSectionHeader(pdf, x, y, pageW, "EXPERIENCIA")
		for _, exp := range experience {
			if exp.Company == "" && exp.Role == "" {
				continue
			}
			// Role (bold) on its own line
			if exp.Role != "" {
				pdf.SetFont("Helvetica", "B", 10)
				pdf.SetXY(x, y)
				pdf.MultiCell(pageW, 5, sanitizeText(exp.Role), "", "L", false)
				y = pdf.GetY() + 0.5
			}
			// Company on its own line
			if exp.Company != "" {
				pdf.SetFont("Helvetica", "", 10)
				pdf.SetXY(x, y)
				pdf.MultiCell(pageW, 5, sanitizeText(exp.Company), "", "L", false)
				y = pdf.GetY() + 0.5
			}
			// Dates (MM/YYYY)
			dateRange := formatDateRange(exp.StartDate, exp.EndDate)
			if dateRange != "" {
				pdf.SetFont("Helvetica", "I", 9)
				pdf.SetXY(x, y)
				pdf.MultiCell(pageW, 4.5, dateRange, "", "L", false)
				y = pdf.GetY() + 0.5
			}
			// Description
			if exp.Description != "" {
				pdf.SetFont("Helvetica", "", 9)
				pdf.SetXY(x, y)
				pdf.MultiCell(pageW, 4.5, sanitizeText(exp.Description), "", "L", false)
				y = pdf.GetY()
			}
			y += 3
		}
		y += 1
	}

	// --- Education ---
	if len(education) > 0 {
		y = drawSectionHeader(pdf, x, y, pageW, "EDUCACAO")
		for _, edu := range education {
			if edu.Institution == "" && edu.Degree == "" {
				continue
			}
			// Degree + Field (bold) on its own line
			left := sanitizeText(edu.Degree)
			if edu.Field != "" {
				left += " - " + sanitizeText(edu.Field)
			}
			if left != "" {
				pdf.SetFont("Helvetica", "B", 10)
				pdf.SetXY(x, y)
				pdf.MultiCell(pageW, 5, left, "", "L", false)
				y = pdf.GetY() + 0.5
			}
			// Institution on its own line
			if edu.Institution != "" {
				pdf.SetFont("Helvetica", "", 10)
				pdf.SetXY(x, y)
				pdf.MultiCell(pageW, 5, sanitizeText(edu.Institution), "", "L", false)
				y = pdf.GetY() + 0.5
			}
			// Dates (MM/YYYY)
			dateRange := formatDateRange(edu.StartDate, edu.EndDate)
			if dateRange != "" {
				pdf.SetFont("Helvetica", "I", 9)
				pdf.SetXY(x, y)
				pdf.MultiCell(pageW, 4.5, dateRange, "", "L", false)
				y = pdf.GetY() + 0.5
			}
			if edu.Notes != "" {
				pdf.SetFont("Helvetica", "", 9)
				pdf.SetXY(x, y)
				pdf.MultiCell(pageW, 4.5, sanitizeText(edu.Notes), "", "L", false)
				y = pdf.GetY()
			}
			y += 3
		}
		y += 1
	}

	// --- Skills ---
	if len(data.Skills) > 0 {
		y = drawSectionHeader(pdf, x, y, pageW, "HABILIDADES")
		pdf.SetFont("Helvetica", "", 10)
		pdf.SetXY(x, y)
		pdf.MultiCell(pageW, 5, sanitizeText(strings.Join(data.Skills, ", ")), "", "L", false)
		y = pdf.GetY() + 4
	}

	// --- Languages ---
	if len(data.Languages) > 0 {
		y = drawSectionHeader(pdf, x, y, pageW, "IDIOMAS")
		pdf.SetFont("Helvetica", "", 10)
		pdf.SetXY(x, y)
		pdf.MultiCell(pageW, 5, sanitizeText(strings.Join(data.Languages, ", ")), "", "L", false)
		y = pdf.GetY() + 4
	}

	// --- Certifications ---
	if len(data.Certifications) > 0 {
		y = drawSectionHeader(pdf, x, y, pageW, "CERTIFICACOES")
		pdf.SetFont("Helvetica", "", 10)
		pdf.SetXY(x, y)
		pdf.MultiCell(pageW, 5, sanitizeText(strings.Join(data.Certifications, ", ")), "", "L", false)
	}

	var buf strings.Builder
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar PDF: %w", err)
	}

	return []byte(buf.String()), nil
}

func sortExperience(items []ExperienceEntry) {
	sort.SliceStable(items, func(i, j int) bool {
		yi, mi := parseDate(items[i].StartDate)
		yj, mj := parseDate(items[j].StartDate)
		if yi != yj {
			return yi > yj
		}
		return mi > mj
	})
}

func sortEducation(items []EducationEntry) {
	sort.SliceStable(items, func(i, j int) bool {
		yi, mi := parseDate(items[i].StartDate)
		yj, mj := parseDate(items[j].StartDate)
		if yi != yj {
			return yi > yj
		}
		return mi > mj
	})
}

func drawSectionHeader(pdf *gofpdf.Fpdf, x, y, w float64, title string) float64 {
	pdf.SetFont("Helvetica", "B", 12)
	pdf.SetTextColor(44, 62, 80)
	pdf.SetXY(x, y)
	pdf.CellFormat(w, 7, title, "", 1, "L", false, 0, "")
	y = pdf.GetY()
	drawLine(pdf, x, y, x+w)
	pdf.SetTextColor(0, 0, 0)
	return y + 3
}

func drawLine(pdf *gofpdf.Fpdf, x1, y1, x2 float64) {
	pdf.SetDrawColor(180, 180, 180)
	pdf.SetLineWidth(0.3)
	pdf.Line(x1, y1, x2, y1)
}

func formatDateRange(start, end string) string {
	start = normalizeDate(start)
	end = normalizeDate(end)
	if start == "" && end == "" {
		return ""
	}
	if start == "" {
		return end
	}
	if start == "Atual" {
		return "Atual"
	}
	if end == "" || end == "Atual" {
		return start + " - Atual"
	}
	return start + " - " + end
}

func normalizeDate(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if isPresentDate(s) {
		return "Atual"
	}
	y, m := parseDate(s)
	if y <= 0 {
		return s
	}
	if m > 0 {
		return fmt.Sprintf("%02d/%d", m, y)
	}
	return strconv.Itoa(y)
}

func isPresentDate(s string) bool {
	switch strings.ToLower(s) {
	case "atual", "presente", "present", "current", "hoje", "today", "ate agora", "até agora":
		return true
	}
	return false
}

func parseDate(s string) (int, int) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return 0, 0
	}

	// ISO: 2020-03-15, 2020/03
	if m := regexp.MustCompile(`(\d{4})[/-](\d{1,2})`).FindStringSubmatch(s); m != nil {
		y, _ := strconv.Atoi(m[1])
		mo, _ := strconv.Atoi(m[2])
		return y, mo
	}

	// BR/other: 15/03/2020, 03/2020, 3-2020
	if m := regexp.MustCompile(`(\d{1,2})[/-](\d{4})`).FindStringSubmatch(s); m != nil {
		mo, _ := strconv.Atoi(m[1])
		y, _ := strconv.Atoi(m[2])
		if mo >= 1 && mo <= 12 {
			return y, mo
		}
		return y, 0
	}

	// Year only (with optional month name)
	if m := regexp.MustCompile(`(\d{4})`).FindStringSubmatch(s); m != nil {
		y, _ := strconv.Atoi(m[1])
		return y, monthFor(s)
	}

	return 0, 0
}

func monthFor(s string) int {
	months := map[string]int{
		"janeiro": 1, "january": 1, "jan": 1,
		"fevereiro": 2, "february": 2, "feb": 2, "fev": 2,
		"março": 3, "marco": 3, "march": 3, "mar": 3,
		"abril": 4, "april": 4, "apr": 4, "abr": 4,
		"maio": 5, "may": 5, "mai": 5,
		"junho": 6, "june": 6, "jun": 6,
		"julho": 7, "july": 7, "jul": 7,
		"agosto": 8, "august": 8, "aug": 8,
		"setembro": 9, "september": 9, "sep": 9, "set": 9,
		"outubro": 10, "october": 10, "oct": 10,
		"novembro": 11, "november": 11, "nov": 11,
		"dezembro": 12, "december": 12, "dec": 12, "dez": 12,
	}
	for name, num := range months {
		if strings.Contains(s, name) {
			return num
		}
	}
	return 0
}

func sanitizeText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = strings.TrimRight(s, "\n")
	return utf8ToCP1252(s)
}

var utf8ToCP1252 = func() func(string) string {
	return gofpdf.New("P", "mm", "A4", "").UnicodeTranslatorFromDescriptor("")
}()
