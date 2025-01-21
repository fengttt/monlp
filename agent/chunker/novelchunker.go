package chunker

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/matrixorigin/monlp/chunk"
)

type NovelChunkerConfig struct {
	StringMode bool `json:"string_mode"`
}

// NovelChunker is a chunker for novels.
type NovelChunkerInputData struct {
	Url string `json:"url"`
}

type NovelChunkerInput struct {
	Data NovelChunkerInputData `json:"data"`
}

type NovelChunkerOutput struct {
	Data []chunk.Chunk `json:"data"`
}

type NovelChunkerStrOutput struct {
	Data [][]string `json:"data"`
}

type NovelChunker struct {
	conf NovelChunkerConfig
}

func (c *NovelChunker) Config(bs []byte) error {
	// unmarshal config
	if bs == nil {
		return nil
	}
	err := json.Unmarshal(bs, &c.conf)
	return err
}

func (c *NovelChunker) Close() error {
	return nil
}

func (c *NovelChunker) Execute(input []byte, dict map[string]string) ([]byte, error) {
	if len(input) == 0 {
		return nil, nil
	}

	// unmarshal input to NovelChunkerInput
	var novelChunkerInput NovelChunkerInput
	err := json.Unmarshal(input, &novelChunkerInput)
	if err != nil {
		return nil, err
	}

	// only handle file:// for now
	if len(novelChunkerInput.Data.Url) < 7 || novelChunkerInput.Data.Url[:7] != "file://" {
		return nil, fmt.Errorf("invalid url: %s", novelChunkerInput.Data.Url)
	}
	// Open the file
	file, err := os.Open(novelChunkerInput.Data.Url[7:])
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read the file, call chunker
	chunks, err := chunk.NewNovelChunker(file)
	if err != nil {
		return nil, err
	}

	// Marshal the output
	if c.conf.StringMode {
		var output NovelChunkerStrOutput
		for chunk := range chunks.Chunk() {
			row := make([]string, 5)
			row[0] = strconv.Itoa(int(chunk.Num1))
			row[1] = strconv.Itoa(int(chunk.Num2))
			row[2] = chunk.Path
			row[3] = chunk.Title
			row[4] = chunk.Text
			output.Data = append(output.Data, row)
		}
		bs, err := json.Marshal(output)
		if err != nil {
			return nil, err
		}
		return bs, nil
	} else {
		var output NovelChunkerOutput
		for chunk := range chunks.Chunk() {
			output.Data = append(output.Data, chunk)
		}
		bs, err := json.Marshal(output)
		if err != nil {
			return nil, err
		}
		return bs, nil
	}
}
