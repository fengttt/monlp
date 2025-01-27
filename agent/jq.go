package agent

import (
	"encoding/json"

	"github.com/itchyny/gojq"
)

// JqInput and JqOutput should be valid json.

type jq struct {
	NilConfigAgent
	NilCloseAgent
	SimpleExecuteAgent
	qstr        string
	parsedQuery *gojq.Query
}

func NewJqAgent(qstr string) (*jq, error) {
	var err error
	var ja jq
	if qstr != "" {
		if err = ja.SetQuery(qstr); err != nil {
			return nil, err
		}
	}
	ja.Self = &ja
	return &ja, nil
}

func (ja *jq) SetQuery(qstr string) error {
	var err error
	ja.qstr = qstr
	ja.parsedQuery, err = gojq.Parse(qstr)
	return err
}

func (ja *jq) ExecuteOne(input []byte, dict map[string]string, yield func([]byte, error) bool) error {
	if ja.parsedQuery == nil {
		if !yield(input, nil) {
			return ErrYieldDone
		}
		return nil
	}

	var jsonInput any
	err := json.Unmarshal(input, &jsonInput)
	if err != nil {
		return err
	}

	iter := ja.parsedQuery.Run(jsonInput)
	for v, ok := iter.Next(); ok; v, ok = iter.Next() {
		if verr, iserr := v.(error); iserr {
			if !yield(nil, verr) {
				return ErrYieldDone
			}
		} else {
			bs, err := json.Marshal(v)
			if err != nil {
				if !yield(nil, err) {
					return ErrYieldDone
				}
			}
			if !yield(bs, nil) {
				return ErrYieldDone
			}
		}
	}
	return nil
}
