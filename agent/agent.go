// agent package defines the interface for the agent, which is the basic unit of the pipeline
// or workflow.   Each agent process input data and produce output data, then user can connect
// multiple agents to form a pipeline.
// The both input and output are json, with an optional desc field and a data field.
//
//	{
//	    "desc": An optional description json object
//	    "data": A json object
//	}
//
// The data and desc json object are agent specific and should have been documented by each
// agent.
package agent

import "fmt"

type Agent interface {
	// Config the agent with a config file
	Config(bs []byte) error
	// Execute the agent with input data, return the output data and error
	Execute(input []byte) ([]byte, error)
}

type AgentPipe struct {
	Agents []Agent
}

func (ap *AgentPipe) Config(bs []byte) error {
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

func (af *AgentFanOut) Config(bs []byte) error {
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
