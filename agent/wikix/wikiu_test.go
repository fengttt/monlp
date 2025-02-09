package wikix

import (
	"testing"

	"github.com/matrixorigin/monlp/common"
	"github.com/matrixorigin/monlp/textu/extract"
	gowiki "github.com/trietmn/go-wiki"
)

func TestWikiUPage(t *testing.T) {
	lines, err := scanQuestionLines()
	common.Assert(t, err == nil, "scanQuestionLines failed: %v", err)
	for _, line := range lines {
		r, _, err := gowiki.Search(line, 3, false)
		common.Assert(t, err == nil, "Expected nil, got %v", err)

		for _, v := range r {
			t.Logf("Page: %s", v)
		}
	}

	page, err := gowiki.GetPage("Michael Freedman", -1, false, true)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	t.Logf("Page: %v", page)

	content, err := page.GetContent()
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	t.Logf("Page Content: %s", content)

	wikitext, err := GetWikiText(page.Title)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	t.Logf("Page WikiText: %s", wikitext)

	// html, err := page.GetHTML()
	// t.Logf("Page HTML: %s", html)

	// off, err := common.CreateFileForTest("data", "mf.html")
	// common.Assert(t, err == nil, "CreateFileForTest failed: %v", err)
	// off.Write([]byte(html))
	// off.Close()

	// links, err := page.GetLink()
	// t.Logf("Page Links: %v", links)
}

func TestWikiText(t *testing.T) {
	page, err := gowiki.GetPage("Michael Freedman", -1, false, true)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	t.Logf("Page: %v", page)

	wikitext, err := GetWikiText(page.Title)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	t.Logf("Page WikiText: %s", wikitext)
}

func TestInfoBox(t *testing.T) {
	wikitext, err := GetWikiText("Jason")
	common.Assert(t, err == nil, "Expected nil, got %v", err)

	var ex extract.WikiInfoBoxExtractor
	for exv := range ex.Extract("Jason", wikitext) {
		t.Logf("Key: %s, Value: %s", exv.Value, exv.Value2)
	}
}
