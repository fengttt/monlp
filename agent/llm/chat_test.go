package llm

import (
	"encoding/json"
	"testing"

	"github.com/matrixorigin/monlp/agent"
	"github.com/matrixorigin/monlp/common"
	"github.com/ollama/ollama/api"
)

func TestSimpleChat(t *testing.T) {
	qs := ChatInput{
		Messages: []api.Message{
			{Role: "user", Content: "1+2="},
			{Role: "user", Content: "What is the color of the sky?"},
			{Role: "user", Content: "What is the capital of France?"},
		},
	}

	qss, err := json.Marshal(&qs)
	common.Assert(t, err == nil, "Expected nil, got %v", err)

	stra := agent.NewStringArrayAgent([]string{
		string(qss),
	})

	chat := NewChatWithPrompt("llama3.2-vision", "You are a helpful assistant.  You should answer the question in one word.", nil)

	var pipe agent.AgentPipe
	pipe.AddAgent(stra)
	pipe.AddAgent(chat)
	defer pipe.Close()

	it, err := pipe.Execute(nil, nil)
	common.Assert(t, err == nil, "Expected nil, got %v", err)

	for data, err := range it {
		common.Assert(t, err == nil, "Expected nil, got %v", err)
		t.Logf("Query Result: %s", string(data))
	}

	chat2 := NewChatWithPrompt("deepseek-r1:14b", "You are a helpful assistant.  You should answer the question in one word.", nil)
	var pipe2 agent.AgentPipe
	pipe2.AddAgent(stra)
	pipe2.AddAgent(chat2)
	defer pipe2.Close()

	it2, err := pipe2.Execute(nil, nil)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	for data, err := range it2 {
		common.Assert(t, err == nil, "Expected nil, got %v", err)
		t.Logf("Query Result: %s", string(data))
	}
}
