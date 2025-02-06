package wikix

import (
	"bufio"
	"encoding/json"
	"testing"

	"github.com/matrixorigin/monlp/agent"
	"github.com/matrixorigin/monlp/agent/llm"
	"github.com/matrixorigin/monlp/common"
)

func scanQuestionLines() ([]string, error) {
	qfile, err := common.OpenFileForTest("agent", "wikix", "questions.txt")
	if err != nil {
		return nil, err
	}
	defer qfile.Close()

	// next read qfile line by line
	var qlines []string
	scanner := bufio.NewScanner(qfile)
	for scanner.Scan() {
		qline := scanner.Text()
		qlines = append(qlines, qline)
	}
	return qlines, nil
}

var (
	testModel = "qwen2.5:14b"
	// testModel = "deepseek-r1:14b"
)

func TestWikiSimple(t *testing.T) {
	qlines, err := scanQuestionLines()
	common.Assert(t, err == nil, "scanQuestionLines failed: %v", err)

	var lines []string
	for _, ql := range qlines {
		line := `{"messages": [{"role": "user", "content": "` + ql + `"}]}`
		lines = append(lines, line)
	}

	stra := agent.NewStringArrayAgent(lines)
	chat := llm.NewChatWithPrompt(testModel, "You are a helpful assistant.  Please answer questions in one sentence.", nil)
	var pipe agent.AgentPipe
	pipe.AddAgent(stra)
	pipe.AddAgent(chat)
	defer pipe.Close()

	cnt := 0
	it, err := pipe.Execute(nil, nil)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	for data, err := range it {
		common.Assert(t, err == nil, "Expected nil, got %v", err)
		var output llm.ChatOutput
		err = json.Unmarshal(data, &output)
		common.Assert(t, err == nil, "Expected nil, got %v", err)

		t.Logf("Question: %s\n", qlines[cnt])
		t.Logf("Answer: %s\n", output.Response.Message.Content)
		cnt++
	}
}

func TestRunTopics(t *testing.T) {
	thisDir := common.ProjectPath("agent", "wikix")
	wikix, err := NewWikiX(thisDir, testModel, SystemPrompt)
	common.Assert(t, err == nil, "NewWikiX failed: %v", err)

	qlines, err := scanQuestionLines()
	common.Assert(t, err == nil, "scanQuestionLines failed: %v", err)

	for _, ql := range qlines {
		wikix.SetValue("userquery", ql)
		topics, err := wikix.runTopics()
		common.Assert(t, err == nil, "RunTopics failed: %v", err)
		t.Logf("Question: %s\n", ql)
		for _, topic := range topics {
			t.Logf("Topic: %s, content: %s\n", topic.Title, topic.Content)
		}
	}
}

func TestRunInitSubq(t *testing.T) {
	thisDir := common.ProjectPath("agent", "wikix")
	wikix, err := NewWikiX(thisDir, testModel, SystemPrompt)
	common.Assert(t, err == nil, "NewWikiX failed: %v", err)

	qlines, err := scanQuestionLines()
	common.Assert(t, err == nil, "scanQuestionLines failed: %v", err)

	for _, ql := range qlines {
		wikix.SetValue("userquery", ql)
		steps, err := wikix.runSubq()
		common.Assert(t, err == nil, "RunTopics failed: %v", err)
		t.Logf("Question: %s\n", ql)
		for _, subq := range steps {
			t.Logf("Subquery: %s\n", subq.Query)
		}
	}
}
