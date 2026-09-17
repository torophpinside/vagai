package services

import (
	"regexp"
	"strings"
)

// TemplateQuestion é o contrato de dados do banco de perguntas-modelo,
// produzindo o mesmo schema do caminho IA (contrato único na API).
type TemplateQuestion struct {
	Category    string `json:"category"`
	Topic       string `json:"topic"`
	Text        string `json:"text"`
	AnswerGuide string `json:"answer_guide"`
}

const (
	maxDetectedTechs        = 6
	techQuestionsPerTech    = 2
	foundationsQuestionAmt  = 4
	architectureQuestionAmt = 3
)

// techDictionary mapeia um termo normalizado (mask) para a tecnologia
// canônica. A ordem define a prioridade de detecção. Casamento com fronteira
// de palavra evita falsos positivos ("governança" não gera "Go").
var techDictionary = []struct{ mask, tech string }{
	{"golang", "Go"}, {"go", "Go"},
	{"gin", "Gin"},
	{"mysql", "MySQL"},
	{"postgresql", "PostgreSQL"}, {"postgres", "PostgreSQL"},
	{"redis", "Redis"},
	{"mongodb", "MongoDB"},
	{"docker", "Docker"},
	{"kubernetes", "Kubernetes"}, {"k8s", "Kubernetes"},
	{"nginx", "Nginx"},
	{"vuejs", "Vue"}, {"vue", "Vue"},
	{"reactjs", "React"}, {"react", "React"},
	{"angular", "Angular"},
	{"typescript", "TypeScript"},
	{"javascript", "JavaScript"},
	{"nodejs", "Node.js"}, {"node", "Node.js"},
	{"python", "Python"},
	{"java", "Java"},
	{"php", "PHP"},
	{"ruby", "Ruby"},
	{"csharp", "C#"}, {"c#", "C#"},
	{"cplusplus", "C++"}, {"c++", "C++"},
	{"dotnet", ".NET"}, {".net", ".NET"},
	{"flutter", "Flutter"},
	{"unity", "Unity"},
	{"gcp", "GCP"}, {"google cloud", "GCP"},
	{"aws", "AWS"}, {"amazon web services", "AWS"},
	{"azure", "Azure"},
	{"git", "Git"},
	{"graphql", "GraphQL"},
	{"kafka", "Kafka"},
	{"rabbitmq", "RabbitMQ"},
	{"elasticsearch", "Elasticsearch"},
	{"terraform", "Terraform"},
	{"tailwind", "Tailwind"},
	{"docker-compose", "Docker Compose"}, {"docker compose", "Docker Compose"},
}

// techRegexes pré-computa as regexes de fronteira de palavra para cada mask,
// evitando reconstrução por chamada de DetectTechnologies.
var techRegexes = func() []struct {
	tech  string
	regex *regexp.Regexp
} {
	out := make([]struct {
		tech  string
		regex *regexp.Regexp
	}, 0, len(techDictionary))
	for _, e := range techDictionary {
		norm := normalizeText(e.mask)
		if norm == "" {
			continue
		}
		re := regexp.MustCompile(`\b` + regexp.QuoteMeta(norm) + `\b`)
		out = append(out, struct {
			tech  string
			regex *regexp.Regexp
		}{tech: e.tech, regex: re})
	}
	return out
}()

// DetectTechnologies extrai tecnologias conhecidas da descrição da vaga,
// sem depender de IA, usando dicionário + texto normalizado. Máximo de 6
// tecnologias, sem duplicatas, caso-insensitive e insensitive a acentos.
func DetectTechnologies(description string) []string {
	norm := normalizeText(description)
	if strings.TrimSpace(norm) == "" {
		return []string{}
	}

	techs := make([]string, 0, maxDetectedTechs)
	seen := make(map[string]bool, maxDetectedTechs)
	for _, e := range techRegexes {
		if len(techs) >= maxDetectedTechs {
			break
		}
		if e.regex.MatchString(norm) {
			if seen[e.tech] {
				continue
			}
			seen[e.tech] = true
			techs = append(techs, e.tech)
		}
	}
	return techs
}

// BuildTemplateQuestions monta o conjunto determinístico de perguntas-modelo:
// até 6 tecnologias (2 perguntas cada) + 4 fundamentos + 3 arquitetura.
// Categorias seguem os enums de models.InterviewCategory.
func BuildTemplateQuestions(techs []string) []TemplateQuestion {
	if len(techs) > maxDetectedTechs {
		techs = techs[:maxDetectedTechs]
	}

	seen := make(map[string]bool, len(techs))
	qs := make([]TemplateQuestion, 0, maxDetectedTechs*techQuestionsPerTech+7)
	for _, tech := range techs {
		if tech == "" || seen[tech] {
			continue
		}
		seen[tech] = true
		qs = append(qs, technologyQuestions(tech)...)
	}

	qs = append(qs, templateFoundationsQuestions()...)
	qs = append(qs, templateArchitectureQuestions()...)
	return qs
}

// extraGeneralQuestions é a reserva de perguntas genéricas usada apenas para
// suplementação (006-fixed-count): quando o conjunto gerado (IA ou template)
// tem menos de TargetQuestionCount itens, estas perguntas completam o total.
// Textos são únicos entre si e em relação ao banco principal.
func extraGeneralQuestions() []TemplateQuestion {
	return []TemplateQuestion{
		{
			Category:    "technology",
			Topic:       "Testes",
			Text:        "Quais tipos de teste você escreve para uma nova funcionalidade e como decide a cobertura ideal?",
			AnswerGuide: "Cubra testes unitários, integração e ponta a ponta; mencione pirâmide de testes, mocks e quando a cobertura vira vaidade.",
		},
		{
			Category:    "technology",
			Topic:       "Git",
			Text:        "Descreva seu fluxo de trabalho com Git em equipe e como você resolve conflitos complexos.",
			AnswerGuide: "Branches por funcionalidade, commits pequenos e descritivos, rebase vs merge, revisão de PR e estratégia de resolução de conflitos.",
		},
		{
			Category:    "technology",
			Topic:       "APIs REST",
			Text:        "O que caracteriza uma API REST bem projetada e quais erros comuns você evita?",
			AnswerGuide: "Recursos nomeados, verbos HTTP corretos, códigos de status, paginação, versionamento e tratamento de erros consistente.",
		},
		{
			Category:    "technology",
			Topic:       "CI/CD",
			Text:        "Como você estrutura um pipeline de CI/CD confiável para um serviço em produção?",
			AnswerGuide: "Stages de lint, teste, build, deploy progressivo, gates de qualidade, rollback automático e observabilidade do pipeline.",
		},
		{
			Category:    "technology",
			Topic:       "Debug",
			Text:        "Descreva sua estratégia para diagnosticar um bug intermitente em produção.",
			AnswerGuide: "Reprodução, logs estruturados, tracing, métricas, hipóteses falsificáveis e correção com teste de regressão.",
		},
		{
			Category:    "technology",
			Topic:       "Segurança",
			Text:        "Quais práticas de segurança você aplica por padrão ao desenvolver uma aplicação web?",
			AnswerGuide: "Validação de entrada, autenticação/autorização, secrets fora do código, dependências atualizadas, headers de segurança e OWASP Top 10.",
		},
		{
			Category:    "foundations",
			Topic:       "Code review",
			Text:        "O que você prioriza ao revisar o código de outra pessoa?",
			AnswerGuide: "Corretude, legibilidade, testes, tratamento de erros e design; feedback objetivo e sugestões acionáveis sem nitpicking.",
		},
		{
			Category:    "foundations",
			Topic:       "Refatoração",
			Text:        "Quando você decide refatorar um trecho de código em vez de apenas fazê-lo funcionar?",
			AnswerGuide: "Sinais como duplicação, funções longas e acoplamento; refatoração incremental com testes como rede de segurança e sem mudar comportamento.",
		},
		{
			Category:    "foundations",
			Topic:       "Complexidade",
			Text:        "Explique o que é complexidade ciclomática e como ela influencia suas decisões de design.",
			AnswerGuide: "Número de caminhos independentes no código; alta complexidade indica necessidade de decomposição, testes extras e simplificação.",
		},
		{
			Category:    "architecture",
			Topic:       "Filas",
			Text:        "Quando você introduz uma fila de mensagens em um sistema e quais trade-offs ela traz?",
			AnswerGuide: "Desacoplamento, absorção de picos e retentativas; custos: consistência eventual, ordenação, duplicação e complexidade operacional.",
		},
		{
			Category:    "architecture",
			Topic:       "Observabilidade",
			Text:        "Quais são os três pilares da observabilidade e como você os usa na prática?",
			AnswerGuide: "Logs, métricas e traces; correlação por request ID, alertas em sintomas (latência, erros, saturação) e dashboards acionáveis.",
		},
		{
			Category:    "architecture",
			Topic:       "Banco de dados",
			Text:        "Como você escolhe entre um banco relacional e um NoSQL para um novo serviço?",
			AnswerGuide: "Considere modelo de dados, necessidade de transações, padrões de acesso, consistência e escala; exemplifique com casos reais.",
		},
	}
}

// SupplementTemplateQuestions retorna perguntas adicionais suficientes para que
// len(existing)+len(returned) alcance target, sem repetir textos já presentes.
// A ordem de preferência é: banco principal (por tecnologia) e, esgotado este,
// a reserva de perguntas genéricas. Nunca retorna mais que o necessário.
func SupplementTemplateQuestions(existing []TemplateQuestion, techs []string, target int) []TemplateQuestion {
	need := target - len(existing)
	if need <= 0 {
		return nil
	}

	seen := make(map[string]bool, len(existing))
	for _, q := range existing {
		seen[q.Text] = true
	}

	out := make([]TemplateQuestion, 0, need)
	take := func(pool []TemplateQuestion) {
		for _, q := range pool {
			if len(out) >= need {
				return
			}
			if seen[q.Text] {
				continue
			}
			seen[q.Text] = true
			out = append(out, q)
		}
	}

	take(BuildTemplateQuestions(techs))
	take(extraGeneralQuestions())
	return out
}

func technologyQuestions(tech string) []TemplateQuestion {
	return []TemplateQuestion{
		{
			Category: "technology",
			Topic:    tech,
			Text:     "Explique quando e por que usar " + tech + " em um projeto, e liste os principais casos de uso.",
			AnswerGuide: "Descreva o problema que " + tech + " resolve, suas vantagens e desvantagens, e dê " +
				"exemplos concretos de cenários em que ela é a escolha mais adequada.",
		},
		{
			Category: "technology",
			Topic:    tech,
			Text:     "Quais são as boas práticas e armadilhas comuns ao trabalhar com " + tech + " em produção?",
			AnswerGuide: "Cite boas práticas reconhecidas (configuração, observabilidade, segurança, performance), " +
				"erros comuns de quem implementa " + tech + " pela primeira vez e como evitá-los.",
		},
	}
}

func templateFoundationsQuestions() []TemplateQuestion {
	return []TemplateQuestion{
		{
			Category: "foundations",
			Topic:    "SOLID",
			Text:     "Explique os princípios SOLID e dê um exemplo prático de onde cada um se aplica.",
			AnswerGuide: "SOLID = Single Responsibility, Open/Closed, Liskov Substitution, Interface Segregation e " +
				"Dependency Inversion. Conecte cada princípio a um exemplo de código real do seu dia a dia.",
		},
		{
			Category: "foundations",
			Topic:    "DRY",
			Text:     "O que significa o princípio DRY e quando ele pode atrapalhar em vez de ajudar?",
			AnswerGuide: "DRY (Don't Repeat Yourself) evita duplicação de conhecimento, não apenas de código. " +
				"Mencione os pontos onde a abstração prematura (evitar repetição cedo demais) pode criar acoplamento indevido.",
		},
		{
			Category: "foundations",
			Topic:    "KISS/YAGNI",
			Text:     "Explique os princípios KISS e YAGNI e como eles guiam decisões de design.",
			AnswerGuide: "KISS (Keep It Simple) prioriza soluções simples; YAGNI (You Aren't Gonna Need It) evita " +
				"implementar funcionalidades especulativas. Relacione com trade-offs de tempo de desenvolvimento e manutenibilidade.",
		},
		{
			Category: "foundations",
			Topic:    "Código limpo",
			Text:     "O que você considera código limpo e quais práticas de revisão você usa para mantê-lo?",
			AnswerGuide: "Nomenclatura clara, funções pequenas, legibilidade, testes e remoção de duplicação. " +
				"Cite práticas de code review e refactoring incremental que você adota.",
		},
	}
}

func templateArchitectureQuestions() []TemplateQuestion {
	return []TemplateQuestion{
		{
			Category: "architecture",
			Topic:    "Design de sistemas",
			Text:     "Projete em alto nível um sistema de 1 milhão de usuários. Cite componentes e trade-offs.",
			AnswerGuide: "Cubra load balancer, API, banco de dados (leitura/escrita), cache, filas, CDN, observabilidade e " +
				"escalabilidade horizontal. Explique trade-offs (consistência vs disponibilidade, custo vs complexidade).",
		},
		{
			Category: "architecture",
			Topic:    "Escalabilidade",
			Text:     "Como você escala um serviço que começa a degradar sob carga crescente?",
			AnswerGuide: "Comece por medição (métricas, tracing, profiling), depois otimize gargalos (N+1, lock, queries), " +
				"cache, filas, réplicas de leitura e, por último, particionamento. Sempre data-driven.",
		},
		{
			Category: "architecture",
			Topic:    "Cache",
			Text:     "Em que situações o cache melhora o sistema e quais riscos você deve gerenciar?",
			AnswerGuide: "Cache é ótimo para leituras frequentes e caras; riscos: invalidação, staleness, cache stampede, " +
				"e distribuição de estado. Discuta TTL, padrões cache-aside e quando não cachear.",
		},
	}
}
