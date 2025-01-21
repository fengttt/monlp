package dbagent

import (
	"encoding/json"
	"testing"

	"github.com/matrixorigin/monlp/agent"
	"github.com/matrixorigin/monlp/common"
)

func TestDbQuery(t *testing.T) {
	// Assume that monlp database is created.
	connstr := ConnStr("localhost", "6001", "dump", "111", "monlp")
	conf := Config{ConnStr: connstr, Table: "testt"}
	// marshal conf to json
	config, err := json.Marshal(conf)
	common.PanicAssert(t, err == nil, "Expected nil, got %v", err)
	qa := DbQuery{}
	err = qa.Config(config)
	common.PanicAssert(t, err == nil, "Expected nil, got %v", err)
	// later will use qa in pipe and it will be closed by pipe.
	// defer qa.Close()

	// test a few queries
	queries := []string{
		"select * from generate_series(1, 3) t",
		"drop table if exists testt",
		"create table testt (a int, b text)",
		"insert into testt values (1, 'a'), (2, 'b')",
		"select * from testt",
	}
	for _, query := range queries {
		out, err := qa.Execute([]byte(`{"data": {"query": "`+query+`"}}`), nil)
		common.Assert(t, err == nil, "Expected nil, got %v", err)
		t.Logf("Query: %s\nResult: %s", query, string(out))
	}

	// pipe a query and writer.
	var pipe agent.AgentPipe
	pipe.AddAgent(&qa)

	wa := DbWriter{}
	wa.Config(config)
	pipe.AddAgent(&wa)

	pipedata := []byte(`{"data": {"query": "select * from testt"}}`)

	out, err := pipe.Execute(pipedata, nil)
	// unmarshal output to DbWriterOutput
	var dbWriterOutput DbWriterOutput
	err = json.Unmarshal(out, &dbWriterOutput)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	common.Assert(t, dbWriterOutput.Data == 2, "Expected 2, got %v", dbWriterOutput.Data)

	out, err = pipe.Execute(pipedata, nil)
	// unmarshal output to DbWriterOutput
	err = json.Unmarshal(out, &dbWriterOutput)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	common.Assert(t, dbWriterOutput.Data == 4, "Expected 4, got %v", dbWriterOutput.Data)

	pipe.Close()
}
