package extract

import (
	"iter"
	"strings"
)

type WikiInfoBoxExtractor struct {
	title     string
	page      string
	lines     []string
	inInfobox bool
}

func (w *WikiInfoBoxExtractor) Extract(title, page string) iter.Seq[Extract] {
	// Extract wiki links, which are in the form [[link]].   We will not
	// handle links across lines for now, not sure if they are possible.
	w.title = title
	w.page = page
	w.lines = strings.Split(page, "\n")

	return func(yield func(Extract) bool) {
		// Find the line {{Infobox
		for _, line := range w.lines {
			if strings.Contains(line, "{{Infobox") {
				w.inInfobox = true
			}

			if !w.inInfobox {
				continue
			} else {
				if strings.HasPrefix(line, "}}") {
					w.inInfobox = false
					break
				}
			}

			// Now I am in the infobox, process line by line
			// Info box is in the form | key = value
			if len(line) > 0 && line[0] == '|' {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0][1:])
					value := strings.TrimSpace(parts[1])
					// strip links from value
					value = strings.ReplaceAll(value, "[[", "")
					value = strings.ReplaceAll(value, "]]", "")

					var ex Extract
					ex.Value = key
					ex.Value2 = value
					if !yield(ex) {
						return
					}
				}
			}
		}
	}
}
