// dbagent implements a database agent that can read or write database.
package dbagent

import "encoding/json"

type Config struct {
	ConnStr string   `json:"connstr"` // connection string
	Table   string   `json:"table"`   // table name
	QTokens []string `json:"qtokens"` // query tokens
}

type DbQueryInputData struct {
	Query string `json:"query"`
}

type DbQueryInput struct {
	Data DbQueryInputData `json:"data"`
}

// Simple and stupid -- everything is a string.
// TODO: support typed columns
type DbQueryOutput struct {
	Data [][]string `json:"data"`
}

type DbQuery struct {
	conf Config
	db   *MoDB
}

func (c *DbQuery) Config(bs []byte) error {
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

func (c *DbQuery) Close() error {
	return c.db.Close()
}

func (c *DbQuery) Execute(input []byte, dict map[string]string) ([]byte, error) {
	if len(input) == 0 {
		return nil, nil
	}

	// unmarshal input to DbQueryInput
	var dbQueryInput DbQueryInput
	err := json.Unmarshal(input, &dbQueryInput)
	if err != nil {
		return nil, err
	}

	rows, err := c.db.Query(dbQueryInput.Data.Query)
	if err != nil {
		return nil, err
	}

	output := DbQueryOutput{Data: rows}
	bs, err := json.Marshal(output)
	if err != nil {
		return nil, err
	}
	return bs, nil
}
