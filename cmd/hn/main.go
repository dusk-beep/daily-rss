package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"time"
)

type Response struct {
	Hits []Hit `json:"hits"`
}

type Hit struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	ObjectID    string `json:"objectID"`
	Author      string `json:"author"`
	Points      int    `json:"points"`
	NumComments int    `json:"num_comments"`
	CreatedAt   string `json:"created_at"`
}

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Items       []RSSItem `xml:"item"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

func main() {
	now := time.Now().UTC()

	// Midnight today UTC.
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	start := today.Add(-24 * time.Hour)
	end := today

	v := url.Values{}
	v.Set("tags", "story")
	v.Set("hitsPerPage", "1000")
	v.Set(
		"numericFilters",
		fmt.Sprintf(
			"created_at_i>%d,created_at_i<%d",
			start.Unix(),
			end.Unix(),
		),
	)

	api := "https://hn.algolia.com/api/v1/search?" + v.Encode()
	resp, err := http.Get(api)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "unexpected status: %s\n", resp.Status)
		os.Exit(1)
	}

	defer resp.Body.Close()

	var r Response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	sort.Slice(r.Hits, func(i, j int) bool {
		return r.Hits[i].Points > r.Hits[j].Points
	})

	if len(r.Hits) > 10 {
		r.Hits = r.Hits[:10]
	}

	var items []RSSItem

	for i, h := range r.Hits {
		link := h.URL
		if link == "" {
			link = "https://news.ycombinator.com/item?id=" + h.ObjectID
		}

		base := time.Now()
		t := base.Add(-time.Duration(i) * time.Minute)

		items = append(items, RSSItem{
			Title:   h.Title,
			Link:    link,
			GUID:    h.ObjectID,
			PubDate: t.Format(time.RFC1123Z),
			Description: fmt.Sprintf(
				"%d points • %d comments • by %s\nhttps://news.ycombinator.com/item?id=%s",
				h.Points,
				h.NumComments,
				h.Author,
				h.ObjectID,
			),
		})
	}

	rss := RSS{
		Version: "2.0",
		Channel: Channel{
			Title:       "Hacker News Yesterday",
			Link:        "https://news.ycombinator.com/",
			Description: "Top Hacker News stories from yesterday (sorted by points)",
			Items:       items,
		},
	}
	var buf bytes.Buffer

	buf.WriteString(xml.Header)

	enc := xml.NewEncoder(&buf)
	enc.Indent("", "  ")
	if err := enc.Encode(rss); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	output := buf.String()
	fmt.Print(output)
}
