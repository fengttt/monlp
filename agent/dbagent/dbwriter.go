package dbagent

import (
	"encoding/json"
	"fmt"
)

type DbWriterInput struct {
	Data [][]string `json:"data"`
}

type DbWriterOutput struct {
	Data int `json:"data"` // number of rows written
}

type DbWriter struct {
	conf Config
	db   *MoDB
}

func (c *DbWriter) Config(bs []byte) error {
	err := json.Unmarshal(bs, &c.conf)
	if err != nil {
		return err
	}

	c.db, err = OpenDB(c.conf.ConnStr)
	if err != nil {
		return err
	}

	if c.conf.Table == "" {
		return fmt.Errorf("Table name is empty")
	}
	return nil
}

func (c *DbWriter) Close() error {
	return c.db.Close()
}

func (c *DbWriter) Execute(input []byte) ([]byte, error) {
	if len(input) == 0 {
		return nil, nil
	}

	var dbWriterInput DbWriterInput
	err := json.Unmarshal(input, &dbWriterInput)
	if err != nil {
		return nil, err
	}

	nRows := len(dbWriterInput.Data)
	if nRows == 0 {
		return nil, nil
	}

	nCols := len(dbWriterInput.Data[0])
	if nCols == 0 {
		return nil, fmt.Errorf("No columns")
	}

	sql := fmt.Sprintf("INSERT INTO %s VALUES (", c.conf.Table)
	for i := 0; i < nCols; i++ {
		if i != 0 {
			sql += ","
		}
		sql += "?"
	}
	sql += ")"
	stmt, err := c.db.Prepare(sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	tx, err := c.db.Begin()
	if err != nil {
		return nil, err
	}

	buf := make([]interface{}, nCols)
	for _, row := range dbWriterInput.Data {
		if len(row) != nCols {
			return nil, fmt.Errorf("Row has %d columns, expected %d", len(row), nCols)
		}
		// copy row to buf, maybe I should quit and use gorm.
		for i, v := range row {
			buf[i] = v
		}
		_, err = stmt.Exec(buf...)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	output := DbWriterOutput{Data: nRows}
	bs, err := json.Marshal(output)
	if err != nil {
		return nil, err
	}
	return bs, nil
}
