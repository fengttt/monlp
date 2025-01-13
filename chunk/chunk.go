package chunk

import (
	"iter"
)

type Chunk struct {
	Num1  int32  // Sequence number of the chunk
	Num2  int32  // Sequence number of the chunk under current path
	Path  string // Path of the chunk
	Title string // Title of the chunk
	Text  string // Text of the chunk
}

type Chunker interface {
	Chunk() iter.Seq[Chunk]
}
