package analytics

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// Constantes de paginação e orçamento de query (research.md #5, SC-020).
const (
	DefaultPage     = 1
	DefaultPageSize = 50
	MaxPageSize     = 100
	queryTimeout    = 5 * time.Second
)

// ErrInvalidPage é retornada para parâmetros de paginação inválidos → HTTP 422
// (contracts/api_contract.md).
var ErrInvalidPage = errors.New("parâmetros de paginação inválidos")

// PageOptions representa a paginação solicitada pelo cliente.
type PageOptions struct {
	Page     int
	PageSize int
}

// ParsePageOptions valida e normaliza page/page_size da query string.
// Padrões: page=1, page_size=50 (máx. 100). Qualquer valor fora do domínio
// (não numérico, <1 ou >100) resulta em ErrInvalidPage.
func ParsePageOptions(pageStr, pageSizeStr string) (PageOptions, error) {
	opts := PageOptions{Page: DefaultPage, PageSize: DefaultPageSize}

	if strings.TrimSpace(pageStr) != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			return opts, ErrInvalidPage
		}
		opts.Page = page
	}

	if strings.TrimSpace(pageSizeStr) != "" {
		size, err := strconv.Atoi(pageSizeStr)
		if err != nil || size < 1 || size > MaxPageSize {
			return opts, ErrInvalidPage
		}
		opts.PageSize = size
	}

	return opts, nil
}

// Offset devolve o OFFSET SQL da opção de paginação.
func (o PageOptions) Offset() int {
	return (o.Page - 1) * o.PageSize
}

// Limit devolve o LIMIT SQL da opção de paginação.
func (o PageOptions) Limit() int {
	return o.PageSize
}

// clampInt limita v ao intervalo [min, max].
func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// medianFloat devolve a mediana de uma fatia ordenada (não modifica a entrada).
func medianFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]float64(nil), values...)
	sortFloats(sorted)
	n := len(sorted)
	if n%2 == 1 {
		return sorted[n/2]
	}
	return (sorted[n/2-1] + sorted[n/2]) / 2
}

// sortFloats insertion-sort leve (bases de analytics são agregados pequenos
// ou já reduzidos pelo banco); evita puxar sort para os vetores grandes.
func sortFloats(v []float64) {
	for i := 1; i < len(v); i++ {
		for j := i; j > 0 && v[j] < v[j-1]; j-- {
			v[j], v[j-1] = v[j-1], v[j]
		}
	}
}

// pct devolve a porcentagem segura (0 quando denominador é 0).
func pct(part, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}

// round2 arredonda para 2 casas decimais (precisão do contrato).
func round2(v float64) float64 {
	if v < 0 {
		return float64(int64(v*100-0.5)) / 100
	}
	return float64(int64(v*100+0.5)) / 100
}

// ratio devolve a razão segura (0 quando denominador é 0).
func ratio(part, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total)
}

// daysBetween devolve a diferença em dias entre dois instantes (0 se inválido).
func daysBetween(from, to time.Time) float64 {
	if to.Before(from) {
		return 0
	}
	return to.Sub(from).Hours() / 24
}
