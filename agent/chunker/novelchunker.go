package chunker

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/matrixorigin/monlp/chunk"
)

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

type NovelChunker struct {
}

func (c *NovelChunker) Config(bs []byte) error {
	return nil
}

func (c *NovelChunker) Execute(input []byte) ([]byte, error) {
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
