package agent

import "fmt"

type Agent interface {
	// Config the agent with a config file
	Config(fn string) error
	// Execute the agent with input data, return the output data and error
	Execute(input []byte) ([]byte, error)
}

type AgentPipe struct {
	Agents []Agent
}

func (ap *AgentPipe) Config(fn string) error {
	return fmt.Errorf("Not implemented")
}

func (ap *AgentPipe) Execute(input []byte) ([]byte, error) {
	var err error
	for _, agent := range ap.Agents {
		input, err = agent.Execute(input)
		if err != nil {
			return nil, err
		}
	}
	return input, nil
}

type AgentFanOut struct {
	Agents []Agent
}

func (af *AgentFanOut) Config(fn string) error {
	return fmt.Errorf("Not implemented")
}

func (af *AgentFanOut) Execute(input []byte) ([]byte, error) {
	var err error
	for _, agent := range af.Agents {
		_, err = agent.Execute(input)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}
