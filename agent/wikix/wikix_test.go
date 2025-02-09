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

func TestWikiSimple(t *testing.T) {
	// pass in -llm deepseek-r1:14b to use different models
	common.ParseFlags([]string{"-vv"})
	qlines, err := scanQuestionLines()
	common.Assert(t, err == nil, "scanQuestionLines failed: %v", err)

	var lines []string
	for _, ql := range qlines {
		line := `{"messages": [{"role": "user", "content": "` + ql + `"}]}`
		lines = append(lines, line)
	}

	stra := agent.NewStringArrayAgent(lines)
	chat := llm.NewChatWithPrompt(common.LLMModel, "You are a helpful assistant.  Please answer questions in one sentence.", nil)
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
	common.ParseFlags([]string{"-vv"})
	thisDir := common.ProjectPath("agent", "wikix")
	wix, err := NewWikiX(thisDir, common.LLMModel, SystemPrompt)
	common.Assert(t, err == nil, "NewWikiX failed: %v", err)

	qlines, err := scanQuestionLines()
	common.Assert(t, err == nil, "scanQuestionLines failed: %v", err)

	for _, ql := range qlines {
		wix.SetValue("userquery", ql)
		err := wix.runTopics()
		common.Assert(t, err == nil, "RunTopics failed: %v", err)
		t.Logf("Question: %s\n", ql)
		for _, topic := range wix.info.Topics {
			shortContent := shortenString(topic.Content, 30)
			t.Logf("Topic: %s (%s), content: %s, ++>> err: %s\n", topic.Title, topic.WikiTitle, shortContent, topic.Err)
		}
	}
}

func TestInitSummarization(t *testing.T) {
	common.ParseFlags([]string{"-vv"})
	thisDir := common.ProjectPath("agent", "wikix")
	wix, err := NewWikiX(thisDir, common.LLMModel, SystemPrompt)
	common.Assert(t, err == nil, "NewWikiX failed: %v", err)

	qlines, err := scanQuestionLines()
	common.Assert(t, err == nil, "scanQuestionLines failed: %v", err)

	// this is slow, so just do the first one.
	wix.SetValue("userquery", qlines[0])
	common.Assert(t, err == nil, "InitSummarization failed: %v", err)
	err = wix.runTopics()
	t.Logf("Question: %s\n", qlines[0])
	for _, topic := range wix.info.Topics {
		shortContent := shortenString(topic.Content, 30)
		t.Logf("Topic: %s (%s), content: %s, err: %s\n", topic.Title, topic.WikiTitle, shortContent, topic.Err)
	}

	err = wix.runSummarize()
	common.Assert(t, err == nil, "RunSummarize failed: %v", err)
	for _, topic := range wix.info.Topics {
		shortContent := shortenString(topic.Summary, 30)
		t.Logf("Topic: %s (%s), Summry: %s, err: %s\n", topic.Title, topic.WikiTitle, shortContent, topic.Err)
	}
}

func TestInitFinal(t *testing.T) {
	common.ParseFlags([]string{"-vvv"})
	thisDir := common.ProjectPath("agent", "wikix")
	wix, err := NewWikiX(thisDir, common.LLMModel, SystemPrompt)
	common.Assert(t, err == nil, "NewWikiX failed: %v", err)

	qlines, err := scanQuestionLines()
	common.Assert(t, err == nil, "scanQuestionLines failed: %v", err)

	for i, ql := range qlines {
		wix.SetValue("userquery", ql)
		err := wix.runTopics()
		common.Assert(t, err == nil || err == ErrNoPageFound, "RunTopics %d failed: %v", i, err)
		err = wix.runFinal()
		common.Assert(t, err == nil, "RunFinal failed: %v", err)
		t.Logf("TestModel: %s, Final Answer at round 0: %s\n", common.LLMModel, wix.info.FinalAnswer)
	}
}

func TestRunInitSubq(t *testing.T) {
	common.ParseFlags([]string{"-vv"})
	thisDir := common.ProjectPath("agent", "wikix")
	wikix, err := NewWikiX(thisDir, common.LLMModel, SystemPrompt)
	common.Assert(t, err == nil, "NewWikiX failed: %v", err)

	qlines, err := scanQuestionLines()
	common.Assert(t, err == nil, "scanQuestionLines failed: %v", err)

	for i, ql := range qlines {
		wikix.SetValue("userquery", ql)
		steps, err := wikix.runSubq()
		common.Assert(t, err == nil, "RunTopics %d failed: %v", i, err)
		t.Logf("Question: %s\n", ql)
		for _, subq := range steps {
			t.Logf("Subquery: %s\n", subq.Query)
		}
	}
}
