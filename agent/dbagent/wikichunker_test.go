package dbagent

import (
	"encoding/json"
	"testing"

	"github.com/matrixorigin/monlp/agent"
	"github.com/matrixorigin/monlp/agent/chunker"
	"github.com/matrixorigin/monlp/common"
)

func TestWikiChunker(t *testing.T) {
	// create table
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
			`k varchar(200) not null primary key, ` +
			`title varchar(200), ` +
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

	// wpipe write databases.
	wikifile := "file://" + common.ProjectPath("data", "enwiki-latest-pages-articles-multistream.xml")
	wstra := agent.NewStringArrayAgent([]string{
		`{"data": {"url": "` + wikifile + `"}}`,
	})

	batchSz := 10
	chunker := chunker.NewWikiChunker(batchSz)

	wa := NewDbWriter()
	waconf := Config{
		ConnStr:   connstr,
		Table:     "testwiki",
		QTemplate: "insert into testwiki values (?, ?, ?, ?)",
	}
	waconfig, err := json.Marshal(waconf)
	common.PanicAssert(t, err == nil, "Expected nil, got %v", err)
	wa.Config(waconfig)

	var wpipe agent.AgentPipe
	wpipe.AddAgent(wstra)
	wpipe.AddAgent(chunker)
	wpipe.AddAgent(wa)
	defer wpipe.Close()

	wit, err := wpipe.Execute(nil, nil)
	common.Assert(t, err == nil, "Expected nil, got %v", err)

	nbatch := 0
	for _, err := range wit {
		common.Assert(t, err == nil, "Expected nil, got %v", err)
		nbatch++
		t.Logf("Write batch %d, each %d pages", nbatch, batchSz)

		if nbatch >= 10 {
			break
		}
	}
}
