package analytics

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/anomalyco/vagai-cli/internal/db"
	"github.com/spf13/cobra"
)

// geoRuneFold mapeia os acentos comuns do português para a letra base
// (T066 — duplicado do módulo da API: a CLI não importa código do servidor).
var geoRuneFold = map[rune]rune{
	'á': 'a', 'à': 'a', 'â': 'a', 'ã': 'a',
	'é': 'e', 'ê': 'e',
	'í': 'i',
	'ó': 'o', 'ô': 'o', 'õ': 'o',
	'ú': 'u',
	'ç': 'c',
}

// geoAccentSQL é a versão SQL de geoRuneFold: pares ordenados de troca
// (REPLACE) para remover acentos do texto das vagas.
var geoAccentSQL = [][2]string{
	{"á", "a"}, {"à", "a"}, {"â", "a"}, {"ã", "a"},
	{"é", "e"}, {"ê", "e"},
	{"í", "i"},
	{"ó", "o"}, {"ô", "o"}, {"õ", "o"},
	{"ú", "u"},
	{"ç", "c"},
}

// geoCityToUF mapeia cidades brasileiras (minúsculas e sem acentos) para a
// UF (T066): jobs não têm coluna de cidade — a localização é inferida por
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

// geoNameCount é a linha genérica das agregações (name + cnt).
type geoNameCount struct {
	Name string
	Cnt  int64
}

// geoRemoteRow guarda os totais brutos da query de trabalho remoto.
type geoRemoteRow struct {
	Total  float64
	Remote float64
}

// geoCityRef é o par nome normalizado → UF de uma cidade do mapa.
type geoCityRef struct {
	Name string
	UF   string
}

var geographicCmd = &cobra.Command{
	Use:   "geographic",
	Short: "Relatório de insights geográficos e demográficos",
	Long: `Usuários por cidade e fuso horário, vagas por cidade e UF
(inferidas do título/descrição) e a razão de vagas amigáveis a
trabalho remoto.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		org, err := mustSetup(cmd)
		if err != nil {
			return err
		}
		printHeader("Geographic & Demographic Insights", org)

		// Estado vazio (T069): zero usuários com membership E zero vagas.
		var users, jobs int64
		if err := db.DB.Raw(`
			SELECT COUNT(DISTINCT u.id)
			FROM users u JOIN memberships ms
				ON ms.user_id = u.id AND ms.organization_id = ?
			WHERE u.deleted_at IS NULL`, org).Scan(&users).Error; err != nil {
			return err
		}
		if err := db.DB.Table("jobs").Where("organization_id = ?", org).Count(&jobs).Error; err != nil {
			return err
		}
		if users == 0 && jobs == 0 {
			fmt.Println("Nenhuma informação geográfica disponível.")
			fmt.Println("Dica: cadastre usuários na organização e colete vagas (crawler) para gerar este relatório.")
			return nil
		}

		// T065 — usuários por cidade (top 10) e por fuso horário (top 10).
		printGeoNameCount("Usuários por cidade (top 10)", `
			SELECT u.city AS name, COUNT(DISTINCT u.id) AS cnt
			FROM users u JOIN memberships ms
				ON ms.user_id = u.id AND ms.organization_id = ?
			WHERE u.deleted_at IS NULL AND u.city <> ''
			GROUP BY u.city ORDER BY cnt DESC LIMIT 10`, org, users)

		printGeoNameCount("Usuários por fuso horário (top 10)", `
			SELECT u.timezone AS name, COUNT(DISTINCT u.id) AS cnt
			FROM users u JOIN memberships ms
				ON ms.user_id = u.id AND ms.organization_id = ?
			WHERE u.deleted_at IS NULL AND u.timezone <> ''
			GROUP BY u.timezone ORDER BY cnt DESC LIMIT 10`, org, users)

		// T066 — vagas por cidade (uma única passada com SUMs condicionais).
		refs := geoCityRefs()
		byCity, byState, err := geoJobsByLocation(org, refs)
		if err != nil {
			return err
		}

		fmt.Println("\nVagas por cidade (top 10):")
		if len(byCity) == 0 {
			fmt.Println("  (sem dados)")
		} else {
			for i, c := range byCity {
				if i >= 10 {
					break
				}
				fmt.Printf("  %-30s %6d  (%5.1f%%)\n", c.City, c.Count, pct64(c.Count, jobs))
			}
		}

		fmt.Println("\nVagas por estado:")
		if len(byState) == 0 {
			fmt.Println("  (sem dados)")
		} else {
			for i, s := range byState {
				if i >= 10 {
					break
				}
				fmt.Printf("  %-4s %6d  (%5.1f%%)\n", s.State, s.Count, pct64(s.Count, jobs))
			}
		}

		// T067 — trabalho remoto.
		var remote geoRemoteRow
		if err := db.DB.Raw(`
			SELECT COUNT(*) AS total,
				COALESCE(SUM(LOWER(CONCAT(COALESCE(title,''),' ',COALESCE(description,'')))
					REGEXP 'remote|remoto|home office|teletrabalho|hibrido remoto'),0) AS remote
			FROM jobs WHERE organization_id = ?`, org).Scan(&remote).Error; err != nil {
			return err
		}
		total, remoteN := int64(remote.Total), int64(remote.Remote)
		onsite := total - remoteN
		if onsite < 0 {
			onsite = 0
		}
		fmt.Println("\nTrabalho remoto:")
		fmt.Printf("  Remoto: %d de %d vagas (%.1f%%)\n", remoteN, total, pct64(remoteN, total))
		fmt.Printf("  Presencial: %d | Razão remoto:presencial: %s\n", onsite, geoRatioString(remoteN, onsite))

		return nil
	},
}

// printGeoNameCount imprime uma seção de agregação name/cnt com percentual.
func printGeoNameCount(title, query string, org uint, total int64) {
	var rows []geoNameCount
	if err := db.DB.Raw(query, org).Scan(&rows).Error; err != nil {
		fmt.Printf("\n%s:\n  (erro: %v)\n", title, err)
		return
	}
	fmt.Printf("\n%s:\n", title)
	if len(rows) == 0 {
		fmt.Println("  (sem dados)")
		return
	}
	for _, r := range rows {
		fmt.Printf("  %-30s %6d  (%5.1f%%)\n", r.Name, r.Cnt, pct64(r.Cnt, total))
	}
}

// geoJobsByLocation varre as vagas uma única vez com SUMs condicionais por
// cidade (T066) e deriva o total por UF a partir das cidades casadas.
func geoJobsByLocation(org uint, refs []geoCityRef) ([]geoCityCount, []geoStateCount, error) {
	q := geoJobsByCitySQL(refs)
	rows, err := db.DB.Raw(q, org).Rows()
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}
	vals := make([]sql.NullFloat64, len(cols))
	ptrs := make([]any, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}

	counts := make([]int64, len(refs))
	if rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return nil, nil, err
		}
		// As colunas c0..cN seguem a ordem de refs (aliases gerados).
		for i := range cols {
			if i >= len(counts) {
				break
			}
			if vals[i].Valid {
				counts[i] = int64(vals[i].Float64)
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	byUF := make(map[string]int64, len(refs))
	byCity := make([]geoCityCount, 0, len(refs))
	for i, ref := range refs {
		if counts[i] <= 0 {
			continue
		}
		byCity = append(byCity, geoCityCount{City: ref.Name, Count: counts[i]})
		byUF[ref.UF] += counts[i]
	}
	sort.Slice(byCity, func(a, b int) bool {
		if byCity[a].Count != byCity[b].Count {
			return byCity[a].Count > byCity[b].Count
		}
		return byCity[a].City < byCity[b].City
	})

	byState := make([]geoStateCount, 0, len(byUF))
	for uf, n := range byUF {
		byState = append(byState, geoStateCount{State: uf, Count: n})
	}
	sort.Slice(byState, func(a, b int) bool {
		if byState[a].Count != byState[b].Count {
			return byState[a].Count > byState[b].Count
		}
		return byState[a].State < byState[b].State
	})
	return byCity, byState, nil
}

// geoCityCount / geoStateCount espelham os DTOs da API para o relatório.
type geoCityCount struct {
	City  string
	Count int64
}

type geoStateCount struct {
	State string
	Count int64
}

// geoCityRefs devolve as cidades do geoCityToUF em ordem estável, com nomes
// normalizados (geoNormalize é idempotente: minúsculas + sem acentos).
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

// geoJobsByCitySQL monta a query de uma única passada: um SUM condicional
// por cidade sobre LOWER(CONCAT(title, description)) sem acentos. Os nomes
// vêm do mapa literal do arquivo (sem % ou _), portanto o LIKE não precisa de
// escape; não há entrada de usuário no SQL.
func geoJobsByCitySQL(refs []geoCityRef) string {
	var b strings.Builder
	b.WriteString("SELECT ")
	haystack := geoSQLNormalize("CONCAT(COALESCE(title,''),' ',COALESCE(description,''))")
	for i, ref := range refs {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "SUM(%s LIKE '%%%s%%') AS c%d", haystack, ref.Name, i)
	}
	b.WriteString(" FROM jobs WHERE organization_id = ?")
	return b.String()
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
// geoRuneFold (idempotente).
func geoNormalize(s string) string {
	return strings.Map(func(r rune) rune {
		if v, ok := geoRuneFold[r]; ok {
			return v
		}
		return r
	}, strings.ToLower(s))
}

// geoSQLNormalize envolve expr em LOWER + REPLACEs que removem acentos,
// tornando o LIKE insensível a acentuação dos dois lados.
func geoSQLNormalize(expr string) string {
	out := "LOWER(" + expr + ")"
	for _, pair := range geoAccentSQL {
		out = fmt.Sprintf("REPLACE(%s, '%s', '%s')", out, pair[0], pair[1])
	}
	return out
}

func init() {
	Cmd.AddCommand(geographicCmd)
}
