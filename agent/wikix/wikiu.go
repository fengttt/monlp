package wikix

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// WikiQueryResult is the result of a query to the Wikipedia API
// It is used to parse the JSON response from the API.
// Just enough for the GetWikiText, not a complete struct
type WikiQueryResult struct {
	Query struct {
		Normalized []struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"normalized"`
		Pages []struct {
			Pageid    int    `json:"pageid"`
			Ns        int    `json:"ns"`
			Title     string `json:"title"`
			Extract   string `json:"extract"`
			Revisions []struct {
				Slots struct {
					Main struct {
						Contentmodel string `json:"contentmodel"`
						Content      string `json:"content"`
					} `json:"main"`
				} `json:"slots"`
			} `json:"revisions"`
		} `json:"pages"`
	} `json:"query"`
}

func GetWikiText(title string) (string, error) {
	args := map[string]string{
		"action":        "query",
		"prop":          "revisions",
		"rvslots":       "*",
		"rvprop":        "content",
		"formatversion": "2",
		"titles":        title,
		"format":        "json",
	}

	res, err := requestWikiApiBody(args)
	if err != nil {
		return "", err
	}

	var result WikiQueryResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return "", err
	}

	if len(result.Query.Pages) == 0 {
		return "", errors.New("No page found")
	}

	return result.Query.Pages[0].Revisions[0].Slots.Main.Content, nil
}

func requestWikiApiBody(args map[string]string) ([]byte, error) {
	// Make new request object
	url := "http://en.wikipedia.org/w/api.php"
	request, err := http.NewRequest("GET", url, nil)
	// Add header
	request.Header.Set("User-Agent", "go-wiki")
	q := request.URL.Query()

	for k, v := range args {
		q.Add(k, v)
	}
	request.URL.RawQuery = q.Encode()

	full_url := request.URL.String()
	fmt.Println("full_url: ", full_url)

	// Make GET request
	client := http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New("unable to fetch the results")
	}

	// Read body
	body, err := io.ReadAll(res.Body)
	return body, err
}
