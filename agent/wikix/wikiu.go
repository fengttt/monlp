package wikix

import (
	"strings"

	"github.com/matrixorigin/monlp/agent/dbagent"
)

func GetWikiPage(db *dbagent.MoDB, title string) (string, error) {
	// Extremely simplest query.
	// Really should turn on fulltext or something.
	//
	// TODO: Wait -- I can cheat to use gowiki.Search
	//
	rows, err := db.Query("select redirect, content from wiki where k = ?", strings.ToLower(title))
	if err != nil {
		return "", err
	}

	// while we use a for loop, we only process one row.
	for _, row := range rows {
		if row[0] != "" {
			return row[1], nil
		} else {
			return GetWikiPage(db, row[1])
		}
	}

	// nothing found
	return "", nil
}
