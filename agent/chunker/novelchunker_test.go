package chunker

import (
	"testing"

	"github.com/matrixorigin/monlp/common"
)

func TestNovelChunker(t *testing.T) {
	// test data
	book1 := "file://" + common.ProjectPath("data", "t8.shakespeare.txt")
	book2 := "file://" + common.ProjectPath("data", "AnimalFarm.txt")
	book3 := "file://" + common.ProjectPath("data", "xyj.txt")

	agent := NovelChunker{}
	// optional
	agent.Config(nil)
	defer agent.Close()

	out1, err := agent.Execute([]byte(`{"data": {"url": "`+book1+`"}}`), nil)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	t.Logf("t8.shakespear.size: %d", len(out1))

	out1x, err := agent.Execute([]byte(`{"data": {"url": "http://www.google.com"}}`), nil)
	common.Assert(t, err != nil, "Expected error, got nil")
	common.Assert(t, out1x == nil, "Expected nil, got %v", out1x)

	out2, err := agent.Execute([]byte(`{"data": {"url": "`+book2+`"}}`), nil)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	t.Logf("AnimalFarm.size: %d", len(out2))

	out2x, err := agent.Execute([]byte(`{"data": {"url": "file://DoesNotExist"}}`), nil)
	common.Assert(t, err != nil, "Expected error, got nil")
	common.Assert(t, out2x == nil, "Expected nil, got %v", out2x)

	out3, err := agent.Execute([]byte(`{"data": {"url": "`+book3+`"}}`), nil)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	t.Logf("xyj.size: %d", len(out3))

	out4, err := agent.Execute(nil, nil)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	common.Assert(t, out4 == nil, "Expected nil, got %v", out4)
}
func TestNovelStrChunker(t *testing.T) {
	// test data
	book1 := "file://" + common.ProjectPath("data", "t8.shakespeare.txt")
	book2 := "file://" + common.ProjectPath("data", "AnimalFarm.txt")
	book3 := "file://" + common.ProjectPath("data", "xyj.txt")
	book4 := "file://" + common.ProjectPath("data", "红楼梦.txt")

	agent := NovelChunker{}
	agent.Config([]byte(`{"string_mode": true}`))
	defer agent.Close()

	out1, err := agent.Execute([]byte(`{"data": {"url": "`+book1+`"}}`), nil)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	t.Logf("t8.shakespear.size: %d", len(out1))

	out1x, err := agent.Execute([]byte(`{"data": {"url": "http://www.google.com"}}`), nil)
	common.Assert(t, err != nil, "Expected error, got nil")
	common.Assert(t, out1x == nil, "Expected nil, got %v", out1x)

	out2, err := agent.Execute([]byte(`{"data": {"url": "`+book2+`"}}`), nil)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	t.Logf("AnimalFarm.size: %d", len(out2))

	out2x, err := agent.Execute([]byte(`{"data": {"url": "file://DoesNotExist"}}`), nil)
	common.Assert(t, err != nil, "Expected error, got nil")
	common.Assert(t, out2x == nil, "Expected nil, got %v", out2x)

	out3, err := agent.Execute([]byte(`{"data": {"url": "`+book3+`"}}`), nil)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	t.Logf("xyj.size: %d", len(out3))

	out4, err := agent.Execute([]byte(`{"data": {"url": "`+book4+`"}}`), nil)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	t.Logf("hlm.size: %d", len(out4))

	outerr, err := agent.Execute(nil, nil)
	common.Assert(t, err == nil, "Expected nil, got %v", err)
	common.Assert(t, outerr == nil, "Expected nil, got %v", out4)
}
