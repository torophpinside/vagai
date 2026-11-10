package crawler

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func docFrom(t *testing.T, html string) *goquery.Document {
	t.Helper()
	d, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestExtractLocation_JSONLD(t *testing.T) {
	d := docFrom(t, `<script type="application/ld+json">{"@type":"JobPosting","jobLocation":{"address":{"addressLocality":"Barueri","addressRegion":"SP","addressCountry":"BR"}}}</script>`)
	loc, model := extractLocation(d)
	if loc != "Barueri, SP, BR" || model != "" {
		t.Fatalf("got %q %q", loc, model)
	}
}

func TestExtractLocation_JSONLDRemote(t *testing.T) {
	d := docFrom(t, `<script type="application/ld+json">{"@graph":[{"@type":"JobPosting","jobLocationType":"TELECOMMUTE"}]}</script>`)
	_, model := extractLocation(d)
	if model != "remote" {
		t.Fatalf("model=%q", model)
	}
}

func TestExtractLocation_LinkedInSelector(t *testing.T) {
	d := docFrom(t, `<span class="topcard__flavor topcard__flavor--bullet"> Curitiba,  Paraná, Brasil </span>`)
	loc, _ := extractLocation(d)
	if loc != "Curitiba, Paraná, Brasil" {
		t.Fatalf("got %q", loc)
	}
}

func TestDetectWorkModel(t *testing.T) {
	cases := map[string]string{"Curitiba (Híbrido)": "hybrid", "Remote - Brazil": "remote", "SP, Presencial": "presencial", "Curitiba": ""}
	for in, want := range cases {
		if got := detectWorkModel(in); got != want {
			t.Errorf("%q: got %q want %q", in, got, want)
		}
	}
}
