package services

import (
	"regexp"
	"sort"
	"strings"
)

type JobLocation struct {
	Type string
	City string
}

var cityAliases = []struct{ alias, canonical string }{
	{"sao paulo", "sao paulo"}, {"sp", "sao paulo"},
	{"rio de janeiro", "rio de janeiro"}, {"rj", "rio de janeiro"},
	{"belo horizonte", "belo horizonte"}, {"bh", "belo horizonte"},
	{"porto alegre", "porto alegre"}, {"poa", "porto alegre"},
	{"florianopolis", "florianopolis"}, {"floripa", "florianopolis"},
	{"curitiba", "curitiba"}, {"cwb", "curitiba"},
	{"campinas", "campinas"},
	{"sorocaba", "sorocaba"},
	{"santos", "santos"},
	{"brasilia", "brasilia"}, {"bsb", "brasilia"},
	{"salvador", "salvador"}, {"ssa", "salvador"},
	{"fortaleza", "fortaleza"}, {"ce", "ceara"},
	{"recife", "recife"}, {"rec", "recife"},
	{"manaus", "manaus"}, {"mao", "manaus"},
	{"londrina", "londrina"},
	{"maringa", "maringa"},
	{"cascavel", "cascavel"},
	{"ponta grossa", "ponta grossa"},
	{"foz do iguacu", "foz do iguacu"},
	{"joinville", "joinville"},
	{"blumenau", "blumenau"},
	{"goiania", "goiania"},
	{"uberlandia", "uberlandia"},
	{"niteroi", "niteroi"},
	{"vitoria", "vitoria"},
	{"belem", "belem"},
	{"cuiaba", "cuiaba"},
	{"campo grande", "campo grande"},
	{"teresina", "teresina"},
	{"sao luis", "sao luis"},
	{"maceio", "maceio"},
	{"joao pessoa", "joao pessoa"},
	{"natal", "natal"},
	{"lisboa", "lisboa"}, {"lisbon", "lisboa"},
	{"porto", "porto"},
	{"parana", "parana"}, {"pr", "parana"},
	{"santa catarina", "santa catarina"}, {"sc", "santa catarina"},
	{"rio grande do sul", "rio grande do sul"}, {"rs", "rio grande do sul"},
	{"minas gerais", "minas gerais"}, {"mg", "minas gerais"},
	{"espirito santo", "espirito santo"}, {"es", "espirito santo"},
	{"bahia", "bahia"},
	{"ceara", "ceara"},
	{"pernambuco", "pernambuco"}, {"pe", "pernambuco"},
	{"distrito federal", "distrito federal"}, {"df", "distrito federal"},
	{"mato grosso do sul", "mato grosso do sul"},
	{"rio grande do norte", "rio grande do norte"},
	{"tocantins", "tocantins"},
	{"maranhao", "maranhao"},
	{"paraiba", "paraiba"},
	{"alagoas", "alagoas"},
	{"piaui", "piaui"},
	{"amazonas", "amazonas"},
	{"para", "para"},
	{"goias", "goias"},
	{"mato grosso", "mato grosso"},
}

var stateOfCity = map[string]string{
	"curitiba": "parana", "londrina": "parana", "maringa": "parana",
	"cascavel": "parana", "ponta grossa": "parana", "foz do iguacu": "parana",
	"florianopolis": "santa catarina", "joinville": "santa catarina", "blumenau": "santa catarina",
	"porto alegre": "rio grande do sul",
	"sao paulo":    "sao paulo", "campinas": "sao paulo", "sorocaba": "sao paulo", "santos": "sao paulo",
	"rio de janeiro": "rio de janeiro", "niteroi": "rio de janeiro",
	"belo horizonte": "minas gerais", "uberlandia": "minas gerais",
	"vitoria":      "espirito santo",
	"salvador":     "bahia",
	"fortaleza":    "ceara",
	"recife":       "pernambuco",
	"brasilia":     "distrito federal",
	"manaus":       "amazonas",
	"belem":        "para",
	"cuiaba":       "mato grosso",
	"campo grande": "mato grosso do sul",
	"goiania":      "goias",
	"natal":        "rio grande do norte",
	"joao pessoa":  "paraiba",
	"maceio":       "alagoas",
	"teresina":     "piaui",
	"sao luis":     "maranhao",
}

func isStateName(canonical string) bool {
	switch canonical {
	case "parana", "santa catarina", "rio grande do sul", "minas gerais",
		"espirito santo", "bahia", "ceara", "pernambuco", "distrito federal",
		"mato grosso", "mato grosso do sul", "goias", "amazonas", "para",
		"rio grande do norte", "tocantins", "maranhao", "paraiba", "alagoas", "piaui":
		return true
	}
	return false
}

type locationPattern struct {
	re        *regexp.Regexp
	canonical string
}

var locationPatterns = buildLocationPatterns()

func buildLocationPatterns() []locationPattern {
	patterns := make([]locationPattern, 0, len(cityAliases))
	for _, entry := range cityAliases {
		re, err := regexp.Compile(`\b` + regexp.QuoteMeta(entry.alias) + `\b`)
		if err != nil {
			continue
		}
		patterns = append(patterns, locationPattern{re: re, canonical: entry.canonical})
	}
	sort.Slice(patterns, func(i, j int) bool {
		stateI, stateJ := isStateName(patterns[i].canonical), isStateName(patterns[j].canonical)
		if stateI != stateJ {
			return !stateI
		}
		return len(patterns[i].canonical) > len(patterns[j].canonical)
	})
	return patterns
}

var workModelMarkers = []struct {
	marker string
	model  string
}{
	{"hibrido", "hybrid"},
	{"hybrid", "hybrid"},
	{"presencial", "presencial"},
	{"on-site", "presencial"},
	{"onsite", "presencial"},
	{"no escritorio", "presencial"},
	{"home office", "remote"},
	{"trabalho remoto", "remote"},
	{"100% remote", "remote"},
	{"remoto", "remote"},
	{"remota", "remote"},
	{"remotas", "remote"},
	{"remotos", "remote"},
	{"remote", "remote"},
}

var locAccentFold = strings.NewReplacer(
	"á", "a", "à", "a", "ã", "a", "â", "a", "ä", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"í", "i", "ì", "i", "î", "i", "ï", "i",
	"ó", "o", "ò", "o", "õ", "o", "ô", "o", "ö", "o",
	"ú", "u", "ù", "u", "û", "u", "ü", "u",
	"ç", "c", "ñ", "n",
)

func stripAccents(s string) string {
	return locAccentFold.Replace(strings.ToLower(s))
}

func AnalyzeJobLocation(description string) JobLocation {
	desc := stripAccents(description)
	loc := JobLocation{Type: "unknown"}

	for _, m := range workModelMarkers {
		if strings.Contains(desc, m.marker) {
			loc.Type = m.model
			break
		}
	}
	for _, p := range locationPatterns {
		if p.re.MatchString(desc) {
			loc.City = p.canonical
			break
		}
	}
	return loc
}

func normalizeUserCity(userCity string) string {
	stripped := stripAccents(strings.TrimSpace(userCity))
	fields := strings.FieldsFunc(stripped, func(r rune) bool {
		return r == ',' || r == '-' || r == ' ' || r == '.'
	})
	aliases := make([]string, 0, len(cityAliases))
	for _, entry := range cityAliases {
		aliases = append(aliases, entry.alias)
	}
	sort.Slice(aliases, func(i, j int) bool {
		return len(aliases[i]) > len(aliases[j])
	})
	for _, alias := range aliases {
		for _, f := range fields {
			if f == alias {
				for _, entry := range cityAliases {
					if entry.alias == alias {
						return entry.canonical
					}
				}
			}
		}
	}
	return stripped
}

func sameCityOrState(jobCity, userCity string) bool {
	user := normalizeUserCity(userCity)
	job := stripAccents(strings.TrimSpace(jobCity))
	if user == "" || job == "" {
		return user == job
	}
	if user == job {
		return true
	}
	if isStateName(job) {
		if userState, ok := stateOfCity[user]; ok {
			return userState == job
		}
	}
	return false
}

// LocationAllowedToMatch aplica a regra de cidade: presencial/híbrida exige a
// mesma cidade configurada (ou a UF do usuário); remota vale para qualquer
// cidade; sem modelo declarado não bloqueia.
func LocationAllowedToMatch(loc JobLocation, userCity string) bool {
	if normalizeUserCity(userCity) == "" {
		return true
	}
	switch loc.Type {
	case "presencial", "hybrid":
		if loc.City == "" {
			return true
		}
		return sameCityOrState(loc.City, userCity)
	default:
		return true
	}
}
