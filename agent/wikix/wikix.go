package wikix

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/matrixorigin/monlp/agent"
	"github.com/matrixorigin/monlp/agent/dbagent"
	"github.com/matrixorigin/monlp/common"
	"github.com/matrixorigin/monlp/textu/extract"
	"github.com/ollama/ollama/api"
	gowiki "github.com/trietmn/go-wiki"
)

// WikiX is the wiki explorer agent.   It take the same input/output
// as the llm chat agent.

type WikixTopic struct {
	Title       string `json:"title"`
	WikiTitle   string `json:"wiki_title"`
	WikiText    string `json:"wiki_text"`
	WikiInfoBox string `json:"wiki_infobox"`
	Content     string `json:"content"`
	Summary     string `json:"summary"`
	Err         string `json:"err"`
}

type WikixSubquery struct {
	Query  string `json:"query"`
	Answer string `json:"answer"`
}

type WikixLink struct {
	Title     string `json:"title"`
	Paragraph string `json:"paragraph"`
}

const (
	SystemPrompt = `You are a helpful assistant.  
You can search a local wikipedia like knowledge base to 
retrieve information and summarization about a wikipedia
article.  The retrieved wikipedia article, which has a 
title and a content.  

You should try to answer user's question by decomposing the 
original question into smaller and simpler steps.  

USE ONLY information retrieved from the local knowledge base.
DO NOT use information from other sources.
DO NOT make up information.
`
)

type WikixInfo struct {
	// Original User query
	UserQuery string `json:"user_query"`
	// Topics to explore
	Topics []WikixTopic `json:"topics"`
	// Subqueries to ask
	Subqueries []WikixSubquery `json:"subqueries"`
	//
	FinalAnswer string `json:"final_answer"`

	// Current question
	Question string `json:"question"`
	// Current aritcle
	Article string `json:"article"`
	// Current links
	Links []WikixLink `json:"links"`
}

func (wi *WikixInfo) clear() {
	*wi = WikixInfo{}
}

func (wi *WikixInfo) GetTopicsString() string {
	buf := &strings.Builder{}
	fmt.Fprintf(buf, "<topics>\n")
	for _, topic := range wi.Topics {
		fmt.Fprintf(buf, "<topic>\n")
		fmt.Fprintf(buf, "<title>%s</title>\n", topic.Title)
		fmt.Fprintf(buf, "<info>\n")
		fmt.Fprintf(buf, "%s", topic.WikiInfoBox)
		fmt.Fprintf(buf, "</info>\n")
		// if we put the content in the xml, it will be too large
		// and llm will not be able to retrieve the infobox.
		// maybe it is better to split the content into smaller
		// chunks.
		fmt.Fprintf(buf, "<content>\n")
		fmt.Fprintf(buf, "%s", topic.Content)
		fmt.Fprintf(buf, "</content>\n")
		fmt.Fprintf(buf, "</topic>\n")
	}
	fmt.Fprintf(buf, "</topics>\n")
	return buf.String()
}

func (wi *WikixInfo) GetSubqueriesString() string {
	buf := &strings.Builder{}
	fmt.Fprintf(buf, "<subqueries>\n")
	for _, q := range wi.Subqueries {
		fmt.Fprintf(buf, "<subquery>\n")
		fmt.Fprintf(buf, "<query>%s</query>\n", q.Query)
		fmt.Fprintf(buf, "<answer>%s</answer>\n", q.Answer)
		fmt.Fprintf(buf, "</subquery>\n")
	}
	fmt.Fprintf(buf, "</subqueries>\n")
	return buf.String()
}

type WikiX struct {
	agent.NilCloseAgent
	agent.NilConfigAgent
	agent.SimpleExecuteAgent

	db *dbagent.MoDB

	info      WikixInfo
	model     string
	sysPrompt string

	dir         string
	topicsTmpl  *template.Template
	linksTmpl   *template.Template
	subqTmpl    *template.Template
	summaryTmpl *template.Template
	finalTmpl   *template.Template
}

func NewWikiX(dir string, model, sysprompt string) (*WikiX, error) {
	ca := &WikiX{
		dir:       dir,
		model:     model,
		sysPrompt: sysprompt,
	}
	ca.Self = ca

	var err error

	driver, connstr := common.DbConnInfoForTest()
	ca.db, err = dbagent.OpenDB(driver, connstr)
	if err != nil {
		return nil, err
	}

	// read a bunch of prompt templates
	ca.topicsTmpl, err = ca.loadTemp("wikix_topics.txt")
	if err != nil {
		return nil, err
	}

	ca.linksTmpl, err = ca.loadTemp("wikix_links.txt")
	if err != nil {
		return nil, err
	}

	ca.subqTmpl, err = ca.loadTemp("wikix_subq.txt")
	if err != nil {
		return nil, err
	}

	ca.summaryTmpl, err = ca.loadTemp("wikix_summary.txt")
	if err != nil {
		return nil, err
	}

	ca.finalTmpl, err = ca.loadTemp("wikix_final.txt")
	if err != nil {
		return nil, err
	}

	return ca, err
}

func (c *WikiX) loadTemp(fn string) (*template.Template, error) {
	fullFn := filepath.Join(c.dir, fn)
	bs, err := os.ReadFile(fullFn)
	if err != nil {
		return nil, err
	}

	t, err := template.New(fn).Parse(string(bs))
	return t, err
}

func (c *WikiX) SetValue(name string, value any) error {
	c.info.clear()
	switch name {
	case "model":
		c.model = value.(string)
	case "sysPrompt":
		c.sysPrompt = value.(string)
	case "userquery":
		c.info.UserQuery = value.(string)
	default:
		return fmt.Errorf("unknown name: %s", name)
	}
	return nil
}

func (c *WikiX) ExecuteOne(input []byte, dict map[string]string, yield func([]byte, error) bool) error {
	return nil
}

func (c *WikiX) chatWithLLM(umsgs []api.Message, fn func(api.ChatResponse) error) error {
	cli, err := api.ClientFromEnvironment()
	if err != nil {
		return err
	}

	req := api.ChatRequest{
		Model:  c.model,
		Stream: new(bool), // stream response default to false
		Options: map[string]interface{}{
			"temperature": 0.0,
		},
	}

	sysmsg := api.Message{
		Role:    "system",
		Content: c.sysPrompt,
	}
	req.Messages = append(req.Messages, sysmsg)
	req.Messages = append(req.Messages, umsgs...)

	ctx := context.Background()
	return cli.Chat(ctx, &req, fn)
}

func (c *WikiX) extractJosnPart(model, data string, dest any) error {
	if true || model == "phi4" {
		str := strings.Split(data, "\n")
		buf := &strings.Builder{}
		output := false
		for _, line := range str {
			if strings.HasPrefix(line, "```") {
				output = !output
				continue
			}
			if output {
				buf.WriteString(line)
				buf.WriteString("\n")
			}
		}
		data = buf.String()
	} else {
		str := strings.Split(data, "\n")
		buf := &strings.Builder{}
		output := false

		// we expect the final out put will be an valid json object or arrary.
		// we expect it ALWAYS begin with a new line with "{" or "[".
		// if strings.HasPrefix(model, "deepseek-r1") {
		//    filter <think></think>
		//
		for _, line := range str {
			if strings.HasPrefix(line, "{") || strings.HasPrefix(line, "[") {
				output = true
			}

			if output {
				buf.WriteString(line)
				buf.WriteString("\n")
			}
		}
		data = buf.String()
	}
	err := json.Unmarshal([]byte(data), dest)
	if err != nil {
		return fmt.Errorf("extractJosnPart: %v", err)
	}
	return nil
}

func (c *WikiX) runTopics() error {
	buf := &strings.Builder{}
	err := c.topicsTmpl.Execute(buf, &c.info)
	if err != nil {
		return err
	}

	content := buf.String()
	umsgs := []api.Message{
		{
			Role:    "user",
			Content: content,
		},
	}

	var topics []WikixTopic

	slog.Debug("RunTopic on", "content", content)
	fn := func(resp api.ChatResponse) error {
		slog.Debug("LLM Response", "msg", resp.Message.Content)
		// break content into lines
		return c.extractJosnPart(c.model, resp.Message.Content, &topics)
	}

	err = c.chatWithLLM(umsgs, fn)
	if err != nil {
		return err
	}

	// build topics in context.
	for _, topic := range topics {
		dup := false
		for _, existing := range c.info.Topics {
			if existing.Title == topic.Title {
				dup = true
				break
			}
		}

		if !dup {
			topic.WikiTitle, err = GetWikiTitle(topic.Title)
			if err != nil {
				return err
			}
			for _, existing := range c.info.Topics {
				if existing.WikiTitle == topic.WikiTitle {
					dup = true
					break
				}
			}

			topic.WikiText, err = GetWikiText(topic.WikiTitle)
			if err != nil {
				return err
			}

			var ex extract.WikiInfoBoxExtractor
			buf := &strings.Builder{}
			for ib := range ex.Extract(topic.WikiTitle, topic.WikiText) {
				fmt.Fprintf(buf, "<entry>\n<name>\n%s\n</name>\n<value>\n%s\n</value>\n</entry>\n", ib.Value, ib.Value2)
			}
			topic.WikiInfoBox = buf.String()

			page, err := gowiki.GetPage(topic.WikiTitle, -1, false, true)
			if err != nil {
				return err
			}
			topic.Content, err = page.GetContent()

			if !dup {
				c.info.Topics = append(c.info.Topics, topic)
			}
		}
	}
	return nil
}

func (c *WikiX) runSummarize() error {
	for i, topic := range c.info.Topics {
		if topic.Summary == "" {
			summary, err := c.summarizeArticle(topic.Content, c.info.UserQuery)
			slog.Debug("summarizeArticle", "summary", summary, "err", err)
			if err != nil {
				c.info.Topics[i].Err = err.Error()
			} else {
				c.info.Topics[i].Summary = summary
			}
		}
	}
	return nil
}

func shortenString(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func (c *WikiX) summarizeArticle(article, query string) (string, error) {
	buf := &strings.Builder{}
	err := c.summaryTmpl.Execute(buf, map[string]string{
		"Article":  article,
		"Question": query,
	})
	if err != nil {
		return "", err
	}

	content := buf.String()
	umsgs := []api.Message{
		{
			Role:    "user",
			Content: content,
		},
	}

	var summaries []string
	ret := &strings.Builder{}

	slog.Debug("summarizeArticle", "article", shortenString(article, 30))
	fn := func(resp api.ChatResponse) error {
		slog.Debug("summarizeArticle", "resp", resp.Message.Content)
		return c.extractJosnPart(c.model, resp.Message.Content, &summaries)
	}

	err = c.chatWithLLM(umsgs, fn)

	for _, summary := range summaries {
		ret.WriteString(summary)
		ret.WriteString("\n")
	}
	result := ret.String()
	return result, err
}

func (c *WikiX) runSubq() ([]WikixSubquery, error) {
	buf := &strings.Builder{}
	err := c.subqTmpl.Execute(buf, &c.info)
	if err != nil {
		return nil, err
	}

	content := buf.String()
	umsgs := []api.Message{
		{
			Role:    "user",
			Content: content,
		},
	}

	var step []WikixSubquery

	fmt.Printf("runSubq: %s\n", content)
	fn := func(resp api.ChatResponse) error {
		fmt.Printf("resp: %v\n", resp.Message.Content)
		return c.extractJosnPart(c.model, resp.Message.Content, &step)
	}

	err = c.chatWithLLM(umsgs, fn)
	return step, err
}

func (c *WikiX) runFinal() error {
	buf := &strings.Builder{}
	err := c.finalTmpl.Execute(buf, &c.info)
	if err != nil {
		return err
	}

	content := buf.String()
	umsgs := []api.Message{
		{
			Role:    "user",
			Content: content,
		},
	}

	slog.Debug("runFinal", "content", content)
	var q WikixSubquery

	fn := func(resp api.ChatResponse) error {
		slog.Debug("runFinal", "resp", resp.Message.Content)
		return c.extractJosnPart(c.model, resp.Message.Content, &q)
	}

	err = c.chatWithLLM(umsgs, fn)
	if err != nil {
		return err
	}

	if q.Answer != "NOT ENOUGH INFORMATION" {
		c.info.FinalAnswer = q.Answer
	}
	return nil
}
