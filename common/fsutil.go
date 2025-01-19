package common

import (
	"fmt"
	"io"
	"os"
)

func OpenUrl(url string) (io.ReadCloser, error) {
	// Only handle file:// for now
	if url[:7] != "file://" {
		return nil, fmt.Errorf("invalid url: %s", url)
	}
	return os.Open(url[7:])
}
