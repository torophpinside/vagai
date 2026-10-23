package utils

import "github.com/microcosm-cc/bluemonday"

// SanitizeHTML remove tags e scripts perigosos, permitindo apenas texto simples.
func SanitizeHTML(input string) string {
	p := bluemonday.StrictPolicy()
	return p.Sanitize(input)
}
