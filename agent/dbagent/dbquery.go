// dbagent implements a database agent that can read or write database.
package dbagent

import (
	"encoding/json"

	"github.com/matrixorigin/monlp/agent"
)

type Config struct {
	ConnStr   string `json:"connstr"`   // connection string
	Table     string `json:"table"`     // table name
	QTemplate string `json:"qtemplate"` // query template
}

type DbQueryInput struct {
	Data string `json:"data"`
}

// Simple and stupid -- everything is a string.
// TODO: support typed columns
type DbQueryOutput struct {
	Data [][]string `json:"data"`
}

type dbQuery struct {
	agent.SimpleExecuteAgent
	conf Config
	db   *MoDB
}

func (c *dbQuery) Config(bs []byte) error {
	err := json.Unmarshal(bs, &c.conf)
	if err != nil {
		return err
	}

	c.db, err = OpenDB(c.conf.ConnStr)
	if err != nil {
		return err
	}
	return nil
}

func (c *dbQuery) Close() error {
	return c.db.Close()
}

func NewDbQuery() *dbQuery {
	ca := &dbQuery{}
	ca.Self = ca
	return ca
}

func (c *dbQuery) ExecuteOne(input []byte, dict map[string]string, yield func([]byte, error) bool) error {
	if len(input) == 0 {
		return nil
	}

	// unmarshal input to DbQueryInput
	var dbQueryInput DbQueryInput
	err := json.Unmarshal(input, &dbQueryInput)
	if err != nil {
		return err
	}

	rows, err := c.db.Query(dbQueryInput.Data)
	if err != nil {
		return err
	}

	output := DbQueryOutput{Data: rows}
	bs, err := json.Marshal(output)
	if !yield(bs, err) {
		return agent.ErrYieldDone
	}
	return nil
}
