package main

import (
	"strings"

	"github.com/abiosoft/ishell/v2"
	"github.com/matrixorigin/monlp/chunk"
	"github.com/matrixorigin/monlp/common"
)

func wikiCmd(c *ishell.Context) {
	if len(c.Args) == 0 {
		c.Println("wiki subcommand is required")
		return
	}

	if c.Args[0] == "stats" {
		// assume the wiki is already downloaded
		f, err := common.OpenFileForTest("data", "enwiki-latest-pages-articles-multistream.xml")
		common.Assert(nil, err == nil, "Failed to open enwiki-latest-pages-articles-multistream.xml, err %v", err)
		defer f.Close()

		chunker, err := chunk.NewWikiChunker(f, false)
		common.Assert(nil, err == nil, "Failed to create wiki chunker, err %v", err)

		nchunk := 0
		redirect := 0
		multirev := 0

		for chunk := range chunker.Chunk() {
			nchunk++
			if chunk.Path != "" {
				redirect++
			}
			if strings.HasPrefix(chunk.Text, "MULTI REV") {
				multirev++
			}
		}

		c.Printf("Total chunks: %d, # of redirects %d, multi rev %d\n", nchunk, redirect, multirev)
	} else {
		c.Printf("Unknown wiki subcommand, %s\n", c.Args[0])
	}
}
