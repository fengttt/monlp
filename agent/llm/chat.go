package llm

//
// A simple chat agent with function calling.
//
import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/matrixorigin/monlp/agent"
	"github.com/ollama/ollama/api"
)

type ChatInput struct {
	Messages []api.Message `json:"messages"`
}

type ChatOutput struct {
	Response api.ChatResponse `json:"response"`
}

type ChatConfig struct {
	Model        string                 `json:"model"`
	SystemPrompt api.Message            `json:"system_prompt"`
	Format       json.RawMessage        `json:"format"`
	Tools        []api.Tool             `json:"tools"`
	Options      map[string]interface{} `json:"options"`
}

type LLMFunctionCall func(api.ToolCall) (string, error)

type chatter struct {
	agent.NilCloseAgent
	agent.SimpleExecuteAgent
	conf     ChatConfig
	req      api.ChatRequest
	toolcall LLMFunctionCall
}

func NewChatWithPrompt(model, sysprompt string, tc LLMFunctionCall) agent.Agent {
	ca := &chatter{}
	ca.conf.Model = model
	ca.conf.SystemPrompt = api.Message{Role: "system", Content: sysprompt}
	ca.toolcall = tc
	ca.Self = ca
	ca.buildRequest()
	return ca
}

func (c *chatter) Config(bs []byte) error {
	// unmarshal config
	if bs == nil {
		return nil
	}
	err := json.Unmarshal(bs, &c.conf)
	c.buildRequest()
	return err
}

func (c *chatter) SetValue(name string, value any) error {
	if name == "tools" {
		tc, ok := value.(func(api.ToolCall) (string, error))
		if !ok {
			return fmt.Errorf("invalid toolcall function")
		}
		c.toolcall = tc
		return nil
	} else if name == "model" {
		c.conf.Model = value.(string)
		c.req.Model = c.conf.Model
		return nil
	}
	return fmt.Errorf("unknown name: %s", name)
}

func (c *chatter) buildRequest() {
	c.req = api.ChatRequest{
		Model:    c.conf.Model,
		Messages: nil,
		Stream:   new(bool), // stream response default to false
		Format:   c.conf.Format,
		Tools:    c.conf.Tools,
		Options:  c.conf.Options,
	}
}

func (c *chatter) ExecuteOne(input []byte, dict map[string]string, yield func([]byte, error) bool) error {
	if len(input) == 0 {
		return nil
	}

	var chatInput ChatInput
	err := json.Unmarshal(input, &chatInput)
	if err != nil {
		return err
	}

	client, err := api.ClientFromEnvironment()
	if err != nil {
		return err
	}

	c.req.Messages = []api.Message{
		c.conf.SystemPrompt,
	}
	c.req.Messages = append(c.req.Messages, chatInput.Messages...)

	ctx := context.Background()

	var output ChatOutput

	for output.Response.Model == "" {
		err = client.Chat(ctx, &c.req, func(resp api.ChatResponse) error {
			if len(resp.Message.ToolCalls) > 0 {
				for _, tc := range resp.Message.ToolCalls {
					if c.toolcall == nil {
						return fmt.Errorf("toolcall function is not set")
					}
					res, err := c.toolcall(tc)
					if err != nil {
						return err
					}

					c.req.Messages = append(c.req.Messages, api.Message{Role: "tool", Content: res})
				}
			} else {
				output.Response = resp
			}
			return nil
		})
	}

	bs, err := json.Marshal(output)
	if !yield(bs, err) {
		return agent.ErrYieldDone
	}
	return nil
}
