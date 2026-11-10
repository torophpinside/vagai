package linkedin

import (
	"net/url"
	"strings"
	"testing"
)

func TestBuildBasic(t *testing.T) {
	got, err := BuildSearchURL(Query{Keywords: "software engineer", Location: "San Francisco, CA"})
	if err != nil {
		t.Fatal(err)
	}
	want := BaseURL + "?keywords=software%20engineer&location=San%20Francisco%2C%20CA"
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestBuildAllFilters(t *testing.T) {
	got, err := BuildSearchURL(Query{
		Keywords: "python developer", Location: "New York, NY",
		DatePosted: PastMonth, ExperienceLevel: Director, RemoteFilter: Remote,
	})
	if err != nil {
		t.Fatal(err)
	}
	u, _ := url.Parse(got)
	q := u.Query()
	for k, v := range map[string]string{
		"keywords": "python developer", "location": "New York, NY",
		"f_TPR": "r2592000", "f_E": "5", "f_WT": "2",
	} {
		if q.Get(k) != v {
			t.Errorf("%s = %q want %q", k, q.Get(k), v)
		}
	}
}

func TestDateCodes(t *testing.T) {
	for d, want := range map[DatePosted]string{Past24Hours: "r86400", PastWeek: "r604800", PastMonth: "r2592000"} {
		got, _ := BuildSearchURL(Query{DatePosted: d})
		if !strings.Contains(got, "f_TPR="+want) {
			t.Errorf("%s: %s", d, got)
		}
	}
}

func TestNoFilters(t *testing.T) {
	got, err := BuildSearchURL(Query{})
	if err != nil || got != BaseURL {
		t.Fatalf("got %q err %v", got, err)
	}
}

func TestEmptyLocationOmitted(t *testing.T) {
	got, _ := BuildSearchURL(Query{Keywords: "go", Location: "  "})
	if strings.Contains(got, "location") {
		t.Fatal(got)
	}
}

func TestSpecialChars(t *testing.T) {
	got, err := BuildSearchURL(Query{Keywords: `developer & analyst "go"`, Location: "São Paulo (SP)"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.ContainsAny(strings.SplitN(got, "?", 2)[1], " \"+") {
		t.Fatalf("unencoded: %s", got)
	}
	u, _ := url.Parse(got)
	if u.Query().Get("keywords") != `developer & analyst "go"` || u.Query().Get("location") != "São Paulo (SP)" {
		t.Fatalf("roundtrip failed: %s", got)
	}
	if err := ValidateURL(got); err != nil {
		t.Fatal(err)
	}
}

func TestCompanyFilters(t *testing.T) {
	got, _ := BuildSearchURL(Query{Keywords: "engineer", CompanyFilters: []string{"Google", "Microsoft"}})
	u, _ := url.Parse(got)
	if u.Query().Get("keywords") != "engineer Google Microsoft" {
		t.Fatal(got)
	}
}

func TestValidationErrors(t *testing.T) {
	cases := map[string]Query{
		"blank keywords": {Keywords: "   "},
		"long keywords":  {Keywords: strings.Repeat("a", 101)},
		"bad date":       {DatePosted: "yesterday"},
		"bad level":      {ExperienceLevel: "god"},
		"bad remote":     {RemoteFilter: "moon"},
		"too many co":    {CompanyFilters: make([]string, 11)},
		"empty co":       {CompanyFilters: []string{""}},
		"long co":        {CompanyFilters: []string{strings.Repeat("a", 51)}},
	}
	for name, q := range cases {
		if _, err := BuildSearchURL(q); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

func TestValidateURL(t *testing.T) {
	if err := ValidateURL("https://example.com/"); err == nil {
		t.Fatal("expected error")
	}
	if err := ValidateURL(BaseURL + "?keywords=go"); err != nil {
		t.Fatal(err)
	}
}
