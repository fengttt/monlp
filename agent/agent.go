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

import (
	"fmt"
	"iter"
)

var (
	ErrExecOneNA   = fmt.Errorf("ExecuteOne not implemented")
	ErrYieldDone   = fmt.Errorf("Yield done")
	ErrStreamBreak = fmt.Errorf("Stream break")
)

type Agent interface {
	// Config the agent with a config file
	Config(bs []byte) error
	// Execute the agent with input data, return the output data and error
	Execute(it iter.Seq2[[]byte, error], dict map[string]string) (iter.Seq2[[]byte, error], error)
	// ExecuteOne
	ExecuteOne(input []byte, dict map[string]string, yield func([]byte, error) bool) error
	// Close the agent
	Close() error
}

type AgentPipe struct {
	agents []Agent
}

func (ap *AgentPipe) Config(bs []byte) error {
	return fmt.Errorf("Not implemented")
}

func (ap *AgentPipe) AddAgent(agent Agent) {
	ap.agents = append(ap.agents, agent)
}

func (ap *AgentPipe) Execute(it iter.Seq2[[]byte, error], dict map[string]string) (iter.Seq2[[]byte, error], error) {
	var err error
	for _, agent := range ap.agents {
		it, err = agent.Execute(it, dict)
		if err != nil {
			return nil, err
		}
	}
	return it, nil
}
func (ap *AgentPipe) ExecuteOne(input []byte, dict map[string]string, yield func([]byte, error) bool) error {
	return ErrExecOneNA
}

func (ap *AgentPipe) Close() error {
	// close ALL, even if we have error in the middle.
	// Only return the first error
	var savedErr error
	for _, agent := range ap.agents {
		err := agent.Close()
		if err != nil {
			savedErr = err
		}
	}
	return savedErr
}

type AgentFanOut struct {
	Agents []Agent
}

func (af *AgentFanOut) Config(bs []byte) error {
	return fmt.Errorf("Not implemented")
}

func (af *AgentFanOut) Execute(input iter.Seq2[[]byte, error], dict map[string]string) (iter.Seq2[[]byte, error], error) {
	var err error
	for _, agent := range af.Agents {
		_, err = agent.Execute(input, dict)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (af *AgentFanOut) ExecuteOne(input []byte, dict map[string]string, yield func([]byte, error) bool) error {
	return ErrExecOneNA
}

func (af *AgentFanOut) Close() error {
	// close ALL, even if we have error in the middle.
	// Only return the first error
	var savedErr error
	for _, agent := range af.Agents {
		err := agent.Close()
		if err != nil {
			savedErr = err
		}
	}
	return savedErr
}

type NilConfigAgent struct {
}

func (na *NilConfigAgent) Config(bs []byte) error {
	return nil
}

type NilCloseAgent struct {
}

func (na *NilCloseAgent) Close() error {
	return nil
}

type SimpleExecuteAgent struct {
	Self Agent
}

func (sa *SimpleExecuteAgent) Execute(input iter.Seq2[[]byte, error], dict map[string]string) (iter.Seq2[[]byte, error], error) {
	return func(yield func([]byte, error) bool) {
		if input == nil {
			return
		}

		for data, err := range input {
			if err != nil {
				yield(nil, err)
				return
			}

			err = sa.Self.ExecuteOne(data, dict, yield)
			if err == ErrYieldDone {
				return
			} else if err != nil {
				yield(nil, err)
			}
		}
	}, nil
}

type stringArrayAgent struct {
	NilConfigAgent
	NilCloseAgent
	values []string
}

func NewStringArrayAgent(values []string) *stringArrayAgent {
	sa := &stringArrayAgent{values: values}
	return sa
}

func (sa *stringArrayAgent) ExecuteOne(input []byte, dict map[string]string, yield func([]byte, error) bool) error {
	return ErrExecOneNA
}

func (sa *stringArrayAgent) Execute(input iter.Seq2[[]byte, error], dict map[string]string) (iter.Seq2[[]byte, error], error) {
	return func(yield func([]byte, error) bool) {
		for _, data := range sa.values {
			if !yield([]byte(data), nil) {
				return
			}
		}
	}, nil
}

func (sa *stringArrayAgent) SetValues(v []string) {
	sa.values = v
}
