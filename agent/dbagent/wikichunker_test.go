package dbagent

import (
	"encoding/json"
	"testing"

	"github.com/matrixorigin/monlp/agent"
	"github.com/matrixorigin/monlp/common"
)

func TestWikiChunker(t *testing.T) {
	connstr := ConnStr("localhost", "6001", "dump", "111", "monlp")
	conf := Config{ConnStr: connstr, Table: "testwiki"}
	config, err := json.Marshal(conf)
	common.PanicAssert(t, err == nil, "Expected nil, got %v", err)

	qa := NewDbQuery()
	err = qa.Config(config)
	common.PanicAssert(t, err == nil, "Expected nil, got %v", err)

	stra := agent.NewStringArrayAgent([]string{
		`{"data": "drop table if exists testwiki"}`,
		`{"data": "create table testwiki (` +
			`title varchar(200) not null primary key, ` +
			`redirect varchar(200), ` +
			`content text)"}`,
	})

	var pipe agent.AgentPipe
	pipe.AddAgent(stra)
	pipe.AddAgent(qa)
	defer pipe.Close()

	it, err := pipe.Execute(nil, nil)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	for _, err := range it {
		common.Assert(t, err == nil, "Expected nil, got %v", err)
	}
}
