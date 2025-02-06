package wikix

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/matrixorigin/monlp/common"
	gowiki "github.com/trietmn/go-wiki"
	"github.com/trietmn/go-wiki/page"
	"github.com/trietmn/go-wiki/utils"
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
	t.Logf("Page Content: %s", content)

	pp, err := GetPageProps(page)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	for k, v := range pp {
		t.Logf("Page Prop: %s = %s", k, v)
	}

	// html, err := page.GetHTML()
	// t.Logf("Page HTML: %s", html)
	// links, err := page.GetLink()
	// t.Logf("Page Links: %v", links)
}

func GetPageProps(page page.WikipediaPage) (map[string]string, error) {
	pageid := strconv.Itoa(page.PageID)
	args := map[string]string{
		"action":      "query",
		"prop":        "info|pageprops",
		"explaintext": "",
		"rvprop":      "ids",
		"titles":      page.Title,
	}
	res, err := utils.WikiRequester(args)
	if err != nil {
		return nil, err
	}
	if res.Error.Code != "" {
		return nil, errors.New(res.Error.Info)
	}

	fmt.Printf("res: %v\n", res)

	return res.Query.Page[pageid].PageProps, nil
}
