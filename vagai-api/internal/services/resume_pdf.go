package services

import (
	"fmt"
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
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	marginLeft := 15.0
	marginRight := 15.0
	pageW, _ := pdf.GetPageSize()
	pageW = pageW - marginLeft - marginRight
	x := marginLeft

	pdf.SetMargins(marginLeft, 10, marginRight)

	// --- Header: Name ---
	pdf.SetFont("Helvetica", "B", 22)
	pdf.SetXY(x, 10)
	pdf.CellFormat(pageW, 10, sanitizeText(data.PersonalInfo.Name), "", 1, "L", false, 0, "")
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
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetXY(x, y)
		pdf.CellFormat(pageW, 5, sanitizeText(strings.Join(contactParts, " | ")), "", 1, "L", false, 0, "")
		y = pdf.GetY() + 4
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
	if len(data.Experience) > 0 {
		y = drawSectionHeader(pdf, x, y, pageW, "EXPERIENCIA")
		for _, exp := range data.Experience {
			if exp.Company == "" && exp.Role == "" {
				continue
			}
			// Role + Company on same line
			pdf.SetFont("Helvetica", "B", 10)
			pdf.SetXY(x, y)
			left := sanitizeText(exp.Role)
			right := sanitizeText(exp.Company)
			pdf.CellFormat(pageW*0.6, 5, left, "", 0, "L", false, 0, "")
			pdf.CellFormat(pageW*0.4, 5, right, "", 1, "R", false, 0, "")
			y = pdf.GetY()

			// Dates
			dateRange := formatDateRange(exp.StartDate, exp.EndDate)
			if dateRange != "" {
				pdf.SetFont("Helvetica", "I", 9)
				pdf.SetXY(x, y)
				pdf.CellFormat(pageW, 4, dateRange, "", 1, "L", false, 0, "")
				y = pdf.GetY()
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
	if len(data.Education) > 0 {
		y = drawSectionHeader(pdf, x, y, pageW, "EDUCACAO")
		for _, edu := range data.Education {
			if edu.Institution == "" && edu.Degree == "" {
				continue
			}
			pdf.SetFont("Helvetica", "B", 10)
			pdf.SetXY(x, y)
			left := sanitizeText(edu.Degree)
			if edu.Field != "" {
				left += " - " + sanitizeText(edu.Field)
			}
			right := sanitizeText(edu.Institution)
			pdf.CellFormat(pageW*0.6, 5, left, "", 0, "L", false, 0, "")
			pdf.CellFormat(pageW*0.4, 5, right, "", 1, "R", false, 0, "")
			y = pdf.GetY()

			dateRange := formatDateRange(edu.StartDate, edu.EndDate)
			if dateRange != "" {
				pdf.SetFont("Helvetica", "I", 9)
				pdf.SetXY(x, y)
				pdf.CellFormat(pageW, 4, dateRange, "", 1, "L", false, 0, "")
				y = pdf.GetY()
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
		pdf.MultiCell(pageW, 5, sanitizeText(strings.Join(data.Skills, " | ")), "", "L", false)
		y = pdf.GetY() + 4
	}

	// --- Languages ---
	if len(data.Languages) > 0 {
		y = drawSectionHeader(pdf, x, y, pageW, "IDIOMAS")
		pdf.SetFont("Helvetica", "", 10)
		pdf.SetXY(x, y)
		pdf.MultiCell(pageW, 5, sanitizeText(strings.Join(data.Languages, " | ")), "", "L", false)
		y = pdf.GetY() + 4
	}

	// --- Certifications ---
	if len(data.Certifications) > 0 {
		y = drawSectionHeader(pdf, x, y, pageW, "CERTIFICACOES")
		pdf.SetFont("Helvetica", "", 10)
		pdf.SetXY(x, y)
		pdf.MultiCell(pageW, 5, sanitizeText(strings.Join(data.Certifications, " | ")), "", "L", false)
	}

	var buf strings.Builder
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar PDF: %w", err)
	}

	return []byte(buf.String()), nil
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
	start = strings.TrimSpace(start)
	end = strings.TrimSpace(end)
	if start == "" && end == "" {
		return ""
	}
	if start == "" {
		return end
	}
	if end == "" {
		return start + " - Presente"
	}
	return start + " - " + end
}

func sanitizeText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = strings.TrimRight(s, "\n")
	return s
}
