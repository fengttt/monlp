package chunker

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/matrixorigin/monlp/agent"
	"github.com/matrixorigin/monlp/chunk"
)

type WikiChunkerInputData struct {
	Url string `json:"url"`
}

type WikiChunkerInput struct {
	Data WikiChunkerInputData `json:"data"`
}

type WikiChunkerOutputData struct {
	Title    string `json:"title"`
	Redirect string `json:"redirect"`
	Content  string `json:"content"`
}

type WikiChunkerOutput struct {
	Data []WikiChunkerOutputData `json:"data"`
}

type wikiChunker struct {
	agent.NilConfigAgent
	agent.NilCloseAgent
	agent.SimpleExecuteAgent
	batchSize int
}

func NewWikiChunker(batchsz int) *wikiChunker {
	ca := &wikiChunker{}
	ca.Self = ca
	ca.batchSize = batchsz
	return ca
}

func (c *wikiChunker) ExecuteOne(input []byte, dict map[string]string, yield func([]byte, error) bool) error {
	if len(input) == 0 {
		return nil
	}

	// unmarshal input to WikiChunkerInput
	var wikiChunkerInput WikiChunkerInput
	err := json.Unmarshal(input, &wikiChunkerInput)
	if err != nil {
		return err
	}

	// only handle file:// for now
	if len(wikiChunkerInput.Data.Url) < 7 || wikiChunkerInput.Data.Url[:7] != "file://" {
		return fmt.Errorf("invalid url: %s", wikiChunkerInput.Data.Url)
	}
	// Open the file
	file, err := os.Open(wikiChunkerInput.Data.Url[7:])
	if err != nil {
		return err
	}
	defer file.Close()

	// create chunker, chunking at page level.
	chunker, err := chunk.NewWikiChunker(file, false)

	npage := 0
	var output WikiChunkerOutput
	for chunk := range chunker.Chunk() {
		output.Data = append(output.Data, WikiChunkerOutputData{
			Title:    chunk.Title,
			Redirect: chunk.Path,
			Content:  chunk.Text,
		})

		npage++
		if npage%c.batchSize == 0 {
			bs, err := json.Marshal(output)
			if err != nil {
				return err
			}
			if !yield(bs, nil) {
				return agent.ErrYieldDone
			}
			output.Data = nil
		}
	}

	// yield the last batch
	if len(output.Data) > 0 {
		bs, err := json.Marshal(output)
		if err != nil {
			return err
		}
		if !yield(bs, nil) {
			return agent.ErrYieldDone
		}
	}
	return nil
}
