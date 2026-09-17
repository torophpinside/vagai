package services

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseDate(t *testing.T) {
	tests := []struct {
		name      string
		in        string
		wantYear  int
		wantMonth int
	}{
		{name: "empty", in: "", wantYear: 0, wantMonth: 0},
		{name: "whitespace", in: "   ", wantYear: 0, wantMonth: 0},
		{name: "iso date", in: "2020-03-15", wantYear: 2020, wantMonth: 3},
		{name: "iso slash month", in: "2020/03", wantYear: 2020, wantMonth: 3},
		{name: "br full date", in: "15/03/2020", wantYear: 2020, wantMonth: 3},
		{name: "month year", in: "03/2020", wantYear: 2020, wantMonth: 3},
		{name: "single digit month", in: "3-2020", wantYear: 2020, wantMonth: 3},
		{name: "year only", in: "2020", wantYear: 2020, wantMonth: 0},
		{name: "pt month name", in: "Março 2020", wantYear: 2020, wantMonth: 3},
		{name: "pt month uppercase", in: "JANEIRO 2020", wantYear: 2020, wantMonth: 1},
		{name: "month after year", in: "2020 dezembro", wantYear: 2020, wantMonth: 12},
		{name: "en month", in: "january 2020", wantYear: 2020, wantMonth: 1},
		{name: "abbrev month", in: "fev 2019", wantYear: 2019, wantMonth: 2},
		{name: "invalid month br", in: "32/2020", wantYear: 2020, wantMonth: 0},
		{name: "garbage", in: "sem data", wantYear: 0, wantMonth: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			y, m := parseDate(tt.in)
			if y != tt.wantYear || m != tt.wantMonth {
				t.Errorf("parseDate(%q) = (%d, %d), want (%d, %d)", tt.in, y, m, tt.wantYear, tt.wantMonth)
			}
		})
	}
}

func TestNormalizeDate(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: ""},
		{name: "spaces", in: "   ", want: ""},
		{name: "presente", in: "presente", want: "Atual"},
		{name: "atual uppercase", in: "ATUAL", want: "Atual"},
		{name: "today", in: "today", want: "Atual"},
		{name: "iso date", in: "2020-03-15", want: "03/2020"},
		{name: "br full date", in: "15/03/2020", want: "03/2020"},
		{name: "month year", in: "03/2020", want: "03/2020"},
		{name: "pt month name", in: "Março 2020", want: "03/2020"},
		{name: "year only", in: "2020", want: "2020"},
		{name: "passthrough", in: "sem data", want: "sem data"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeDate(tt.in); got != tt.want {
				t.Errorf("normalizeDate(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestIsPresentDate(t *testing.T) {
	trueInputs := []string{"atual", "presente", "present", "current", "hoje", "today", "ate agora", "até agora", "ATUAL"}
	for _, in := range trueInputs {
		if !isPresentDate(in) {
			t.Errorf("isPresentDate(%q) = false, want true", in)
		}
	}

	falseInputs := []string{"", "2020", "março", "atual 2020", "presente permanente"}
	for _, in := range falseInputs {
		if isPresentDate(in) {
			t.Errorf("isPresentDate(%q) = true, want false", in)
		}
	}
}

func TestFormatDateRange(t *testing.T) {
	tests := []struct {
		name  string
		start string
		end   string
		want  string
	}{
		{name: "both empty", start: "", end: "", want: ""},
		{name: "only end", start: "", end: "2020-01-01", want: "01/2020"},
		{name: "present start", start: "presente", end: "", want: "Atual"},
		{name: "empty end", start: "2020-03-15", end: "", want: "03/2020 - Atual"},
		{name: "present end", start: "03/2020", end: "presente", want: "03/2020 - Atual"},
		{name: "month year both", start: "03/2020", end: "12/2020", want: "03/2020 - 12/2020"},
		{name: "mixed formats", start: "2021-01-10", end: "presente", want: "01/2021 - Atual"},
		{name: "pt month names", start: "Março 2016", end: "dezembro 2017", want: "03/2016 - 12/2017"},
		{name: "year only both", start: "2020", end: "2022", want: "2020 - 2022"},
		{name: "unparseable start", start: "sem data", end: "", want: "sem data - Atual"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatDateRange(tt.start, tt.end); got != tt.want {
				t.Errorf("formatDateRange(%q, %q) = %q, want %q", tt.start, tt.end, got, tt.want)
			}
		})
	}
}

func TestSanitizeText(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "accent cp1252", in: "João da Silva", want: "Jo\xe3o da Silva"},
		{name: "crlf", in: "a\r\nb", want: "a\nb"},
		{name: "cr", in: "a\rb", want: "a\nb"},
		{name: "trailing newline", in: "abc\n", want: "abc"},
		{name: "unmappable rune", in: "😂", want: "."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeText(tt.in); got != tt.want {
				t.Errorf("sanitizeText(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSortExperience(t *testing.T) {
	items := []ExperienceEntry{
		{Company: "SemData2", StartDate: ""},
		{Company: "Atual", Role: "Sênior", StartDate: "2021-01-10"},
		{Company: "Estagio", Role: "Estagiário", StartDate: "Março 2016"},
		{Company: "SemData1", StartDate: "   "},
		{Company: "Antiga", Role: "Pleno", StartDate: "2018/03"},
	}
	sortExperience(items)

	wantOrder := []string{"Atual", "Antiga", "Estagio", "SemData2", "SemData1"}
	for i, want := range wantOrder {
		if items[i].Company != want {
			t.Fatalf("order[%d] = %q, want %q (all: %+v)", i, items[i].Company, want, companies(items))
		}
	}
}

func TestSortEducation(t *testing.T) {
	items := []EducationEntry{
		{Institution: "USP", Degree: "Bacharelado", StartDate: "02/2012"},
		{Institution: "PUC", Degree: "Pós-graduação", StartDate: "2024-01"},
		{Institution: "FGV", Degree: "MBA", StartDate: "setembro 2024"},
		{Institution: "SemData", Degree: "Curso", StartDate: ""},
	}
	sortEducation(items)

	wantOrder := []string{"FGV", "PUC", "USP", "SemData"}
	for i, want := range wantOrder {
		if items[i].Institution != want {
			t.Fatalf("order[%d] = %q, want %q (all: %+v)", i, items[i].Institution, want, insts(items))
		}
	}
}

func TestGenerateResumePDF_ATSFormat(t *testing.T) {
	data := ResumeData{
		PersonalInfo: PersonalInfo{
			Name:     "João da Silva",
			Email:    "joao@email.com",
			Phone:    "(11) 99999-9999",
			Location: "São Paulo, SP",
			Linkedin: "linkedin.com/in/joaosilva",
			Website:  "joaosilva.dev",
		},
		Summary: "Engenheiro de software com foco em backend Go.",
		Experience: []ExperienceEntry{
			{Company: "Empresa Antiga", Role: "Engenheiro de Software", StartDate: "2018/03", EndDate: "2020/12", Description: "Manutenção de sistemas legados."},
			{Company: "Empresa Atual", Role: "Engenheiro Sênior", StartDate: "2021-01-10", EndDate: "presente", Description: "Liderança técnica em Go."},
			{Company: "Início", Role: "Estagiário", StartDate: "Março 2016", EndDate: "dezembro 2017", Description: "Suporte ao time."},
			{Company: "", Role: "", StartDate: "", EndDate: "", Description: "FANTASMA-NAO-DEVE-APARECER"},
			{Company: "SóEmpresa", Role: "", StartDate: "", EndDate: "", Description: ""},
			{Company: "", Role: "SóCargo", StartDate: "", EndDate: "", Description: ""},
		},
		Education: []EducationEntry{
			{Degree: "Bacharelado", Field: "Ciência da Computação", Institution: "USP", StartDate: "02/2012", EndDate: "12/2016", Notes: "ICMC"},
			{Degree: "Pós-graduação", Field: "Engenharia de Software", Institution: "PUC", StartDate: "2024-01", EndDate: ""},
			{Institution: "SóInstituicao", Degree: "", StartDate: "", EndDate: "", Notes: ""},
			{Institution: "", Degree: "", StartDate: "", EndDate: "", Notes: ""},
		},
		Skills:         []string{"Go", "Python", "PostgreSQL"},
		Languages:      []string{"Português nativo", "Inglês"},
		Certifications: []string{"AWS", "CKA"},
	}

	out, err := GenerateResumePDF(data)
	if err != nil {
		t.Fatalf("GenerateResumePDF() error = %v", err)
	}
	if len(out) == 0 {
		t.Fatal("GenerateResumePDF() returned empty bytes")
	}
	if !strings.HasPrefix(string(out), "%PDF") {
		t.Errorf("expected PDF header, got prefix %q", string(out[:min(8, len(out))]))
	}

	path := filepath.Join(t.TempDir(), "curriculo.pdf")
	if err := os.WriteFile(path, out, 0o644); err != nil {
		t.Fatalf("write pdf: %v", err)
	}
	text, err := extractFromPDF(path)
	if err != nil {
		t.Fatalf("extractFromPDF() error = %v", err)
	}

	// Acentos devem extrair corretamente (CP1252).
	for _, want := range []string{"João da Silva", "São Paulo, SP", "Liderança técnica em Go", "Português nativo", "SóEmpresa"} {
		if !strings.Contains(text, want) {
			t.Errorf("PDF text missing %q", want)
		}
	}

	// Layout ATS: dados em linhas próprias, seções presentes.
	for _, want := range []string{"RESUMO", "EXPERIENCIA", "EDUCACAO", "HABILIDADES", "IDIOMAS", "CERTIFICACOES"} {
		if !strings.Contains(text, want) {
			t.Errorf("PDF text missing section %q", want)
		}
	}

	// Datas normalizadas para MM/YYYY.
	for _, want := range []string{"01/2021 - Atual", "03/2018 - 12/2020", "03/2016 - 12/2017", "02/2012 - 12/2016", "01/2024 - Atual"} {
		if !strings.Contains(text, want) {
			t.Errorf("PDF text missing normalized date %q", want)
		}
	}

	// Habilidades separadas por vírgula e contato unido por " | ".
	if !strings.Contains(text, "Go, Python, PostgreSQL") {
		t.Errorf("skills should be comma-separated, got: %q", text)
	}
	if !strings.Contains(text, "joao@email.com | (11) 99999-9999") {
		t.Errorf("contact line expected joined by ' | '")
	}

	// Ordem reverse-cronológica.
	idxSenior := strings.Index(text, "Engenheiro Sênior")
	idxPleno := strings.Index(text, "Engenheiro de Software")
	idxEstagio := strings.Index(text, "Estagiário")
	if idxSenior == -1 || idxPleno == -1 || idxEstagio == -1 ||
		idxPleno == strings.Index(text, "Engenheiro de software com foco") {
		t.Fatalf("missing/ambiguous experience roles in text: %q", text)
	}
	if !(idxSenior < idxPleno && idxPleno < idxEstagio) {
		t.Errorf("experience not reverse-chronological: senior=%d pleno=%d estagio=%d", idxSenior, idxPleno, idxEstagio)
	}

	idxPos := strings.Index(text, "Pós-graduação")
	idxBach := strings.Index(text, "Bacharelado")
	if idxPos == -1 || idxBach == -1 {
		t.Fatalf("missing education in text: %q", text)
	}
	if !(idxPos < idxBach) {
		t.Errorf("education not reverse-chronological: pos=%d bach=%d", idxPos, idxBach)
	}

	// Entrada sem cargo e sem empresa deve ser ignorada.
	if strings.Contains(text, "FANTASMA-NAO-DEVE-APARECER") {
		t.Errorf("experience without role/company should be skipped")
	}
}

func TestGenerateResumePDF_Minimal(t *testing.T) {
	out, err := GenerateResumePDF(ResumeData{PersonalInfo: PersonalInfo{Name: "Ana"}})
	if err != nil {
		t.Fatalf("GenerateResumePDF() error = %v", err)
	}
	if len(out) == 0 || !strings.HasPrefix(string(out), "%PDF") {
		t.Fatalf("expected valid PDF, got len=%d prefix=%q", len(out), string(out[:min(8, len(out))]))
	}

	path := filepath.Join(t.TempDir(), "min.pdf")
	if err := os.WriteFile(path, out, 0o644); err != nil {
		t.Fatalf("write pdf: %v", err)
	}
	text, err := extractFromPDF(path)
	if err != nil {
		t.Fatalf("extractFromPDF() error = %v", err)
	}
	if !strings.Contains(text, "Ana") {
		t.Errorf("PDF text missing name, got %q", text)
	}
}

func TestExtractFromPDF_NonExistent(t *testing.T) {
	if _, err := extractFromPDF(filepath.Join(t.TempDir(), "nope.pdf")); err == nil {
		t.Error("expected error for non-existent pdf, got nil")
	}
}

func companies(items []ExperienceEntry) []string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.Company+":"+it.StartDate)
	}
	return out
}

func insts(items []EducationEntry) []string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.Institution+":"+it.StartDate)
	}
	return out
}
