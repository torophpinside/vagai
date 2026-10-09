package analytics

// Envelope de resposta padrão dos endpoints de analytics
// (contracts/api_contract.md). Toda resposta carrega:
//   - os campos de dados próprios do endpoint (top-level);
//   - "pagination" (quando aplicável);
//   - "empty_state" (quando não há dados).

// PaginationMeta espelha o bloco "pagination" do contrato.
type PaginationMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
	TotalItems int64 `json:"total_items"`
}

// EmptyState espelha o bloco "empty_state" do contrato. Quando empty=true a
// resposta também traz mensagem de orientação (spec clarification: estado
// vazio = placeholder + guidance).
type EmptyState struct {
	Empty    bool   `json:"empty"`
	Message  string `json:"message"`
	Guidance string `json:"guidance"`
}

// NewPagination calcula a metainformação de paginação.
func NewPagination(opts PageOptions, total int64) PaginationMeta {
	totalPages := 0
	if opts.PageSize > 0 && total > 0 {
		totalPages = int((total + int64(opts.PageSize) - 1) / int64(opts.PageSize))
	}
	return PaginationMeta{
		Page:       opts.Page,
		PageSize:   opts.PageSize,
		TotalPages: totalPages,
		TotalItems: total,
	}
}

// Envelope monta a resposta final: dados + paginação opcional + empty state
// opcional.
func Envelope(data map[string]any, pagination *PaginationMeta, empty *EmptyState) map[string]any {
	out := make(map[string]any, len(data)+2)
	for k, v := range data {
		out[k] = v
	}
	if pagination != nil {
		out["pagination"] = pagination
	}
	if empty != nil {
		out["empty_state"] = empty
	}
	return out
}

// OKEnvelope monta uma resposta com dados e paginação (cenário com dados).
func OKEnvelope(data map[string]any, pagination *PaginationMeta) map[string]any {
	return Envelope(data, pagination, nil)
}

// EmptyEnvelope monta a resposta de estado vazio: dados zerados + orientação.
// Os campos de dados são mantidos para preservar o shape do contrato.
func EmptyEnvelope(data map[string]any, message, guidance string) map[string]any {
	return Envelope(data, nil, &EmptyState{
		Empty:    true,
		Message:  message,
		Guidance: guidance,
	})
}
