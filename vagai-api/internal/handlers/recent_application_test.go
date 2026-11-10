package handlers

import "testing"

func TestNormalizeCompany(t *testing.T) {
	cases := map[string]string{
		"Acme Ltda":            "acme",
		"  ACME  ":             "acme",
		"Açaí Tecnologia S.A.": "acai tecnologia",
		"Acme, Inc.":           "acme",
		"Acme Corp Ltda":       "acme",
		"Foo-Bar  Labs":        "foo bar labs",
		"":                     "",
	}
	for in, want := range cases {
		if got := normalizeCompany(in); got != want {
			t.Errorf("normalizeCompany(%q) = %q, want %q", in, got, want)
		}
	}
	if normalizeCompany("Acme") == normalizeCompany("Acme Labs") {
		t.Error("different companies must not match")
	}
}
