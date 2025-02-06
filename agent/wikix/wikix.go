package wikix

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/matrixorigin/monlp/agent"
	"github.com/matrixorigin/monlp/agent/dbagent"
	"github.com/matrixorigin/monlp/common"
	"github.com/ollama/ollama/api"
)

// WikiX is the wiki explorer agent.   It take the same input/output
// as the llm chat agent.

type WikixTopic struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Summary string `json:"summary"`
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
	fmt.Fprintf(buf, "===== List of Retrieved Topics =====\n")
	for _, topic := range wi.Topics {
		fmt.Fprintf(buf, "    Topic %s:\n", topic.Title)
		fmt.Fprintf(buf, "    %s\n\n", topic.Title)
	}
	return buf.String()
}

func (wi *WikixInfo) GetSubqueriesString() string {
	buf := &strings.Builder{}
	fmt.Fprintf(buf, "===== Queries and Answers from previous Steps =====\n")
	for _, q := range wi.Subqueries {
		fmt.Fprintf(buf, "    Query: %s\n", q.Query)
		fmt.Fprintf(buf, "    Answer:%s\n\n", q.Answer)
	}
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

func (c *WikiX) runTopics() ([]WikixTopic, error) {
	buf := &strings.Builder{}
	err := c.topicsTmpl.Execute(buf, &c.info)
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

	var topics []WikixTopic

	fmt.Printf("runTopic: %s\n", content)
	fn := func(resp api.ChatResponse) error {
		fmt.Printf("resp: %v\n", resp.Message.Content)
		err := json.Unmarshal([]byte(resp.Message.Content), &topics)
		return err
	}

	err = c.chatWithLLM(umsgs, fn)
	return topics, err
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
		err := json.Unmarshal([]byte(resp.Message.Content), &step)
		return err
	}

	err = c.chatWithLLM(umsgs, fn)
	return step, err
}

func (c *WikiX) retrievePages(topic string) (string, error) {
	//
	return "", nil
}
