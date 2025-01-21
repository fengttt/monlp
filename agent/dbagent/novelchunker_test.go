package dbagent

import (
	"encoding/json"
	"testing"

	"github.com/matrixorigin/monlp/agent"
	"github.com/matrixorigin/monlp/agent/chunker"
	"github.com/matrixorigin/monlp/common"
)

func TestLoadNovelChunker(t *testing.T) {
	// open database, create table
	connstr := ConnStr("localhost", "6001", "dump", "111", "monlp")
	conf := Config{ConnStr: connstr, Table: "testnovel"}
	config, err := json.Marshal(conf)
	common.PanicAssert(t, err == nil, "Expected nil, got %v", err)
	qa := DbQuery{}
	err = qa.Config(config)
	common.PanicAssert(t, err == nil, "Expected nil, got %v", err)
	defer qa.Close()

	_, err = qa.Execute([]byte(`{"data": {"query": "drop table if exists testnovel"}}`), nil)
	common.Assert(t, err == nil, "Expected nil, got %v", err)

	// create table
	createTableQuery := `create table testnovel (
		url text not null,
		num1 int not null,
		num2 int not null,
		path text,
		title text,
		content text,
		primary key (url, num1, num2)
	)`
	_, err = qa.Execute([]byte(`{"data": {"query": "`+createTableQuery+`"}}`), nil)
	common.Assert(t, err == nil, "Expected nil, got %v", err)

	// pipeline, chunk novels and write to database.
	ca := chunker.NovelChunker{}
	ca.Config([]byte(`{"string_mode": true}`))

	wa := DbWriter{}
	waconf := Config{
		ConnStr: connstr,
		Table:   "testnovel",
		QTokens: []string{
			"insert into testnovel values (",
			":URL:,",
			" ?, ?, ?, ?, ?)",
		},
	}
	waconfig, err := json.Marshal(waconf)
	common.PanicAssert(t, err == nil, "Expected nil, got %v", err)
	wa.Config(waconfig)

	var pipe agent.AgentPipe
	pipe.AddAgent(&ca)
	pipe.AddAgent(&wa)
	defer pipe.Close()

	// Insert animal farm chunks
	book := "file://" + common.ProjectPath("data", "AnimalFarm.txt")
	dict := make(map[string]string)
	dict["URL"] = "'AnimalFarm'" // single quote for string in sql
	out, err := pipe.Execute([]byte(`{"data": {"url": "`+book+`"}}`), dict)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	t.Logf("AnimalFarm.size: %v", out)

	// Insert xyj chunks
	book = "file://" + common.ProjectPath("data", "xyj.txt")
	dict["URL"] = "'xyj'" // single quote for string in sql
	out, err = pipe.Execute([]byte(`{"data": {"url": "`+book+`"}}`), dict)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	t.Logf("XYJ.size: %v", out)
}
