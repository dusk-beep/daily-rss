package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"strings"

	xhtml "golang.org/x/net/html"
)

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`

	Solution struct {
		Status   int    `json:"status"`
		Response string `json:"response"`
	} `json:"solution"`
}

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Language    string `xml:"language,omitempty"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description,omitempty"`
	PubDate     string `xml:"pubDate,omitempty"`
	GUID        string `xml:"guid,omitempty"`
	Category    string `xml:"category,omitempty"`
}

func extractXML(data string) (string, error) {
	doc, err := xhtml.Parse(strings.NewReader(data))
	if err != nil {
		return "", err
	}

	var pre *xhtml.Node
	var f func(*xhtml.Node)
	f = func(n *xhtml.Node) {
		if n.Type == xhtml.ElementNode && n.Data == "pre" {
			pre = n
			return
		}
		for c := n.FirstChild; c != nil && pre == nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	if pre == nil {
		return "", fmt.Errorf("no <pre> found")
	}

	var b strings.Builder
	for c := pre.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == xhtml.TextNode {
			b.WriteString(c.Data)
		}
	}

	xm := b.String()
	if xm == "" {
		return "", fmt.Errorf("<pre> contains no XML")
	}
	return xm, nil
}

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "Usage: %s <response.json> <ziperto.html> <output.xml>\n", os.Args[0])
		os.Exit(2)
	}

	responsePath := os.Args[1]
	htmlPath := os.Args[2]
	outputPath := os.Args[3]

	data, err := os.ReadFile(responsePath)
	if err != nil {
		panic(err)
	}

	var r Response
	if err := json.Unmarshal(data, &r); err != nil {
		panic(err)
	}

	if r.Status != "ok" {
		panic(fmt.Errorf("flaresolverr error: %s", r.Message))
	}

	if r.Solution.Status != 200 {
		panic(fmt.Errorf("unexpected HTTP status: %d", r.Solution.Status))
	}

	xm, err := extractXML(r.Solution.Response)
	if err != nil {
		panic(err)
	}

	var rss RSS
	if err := xml.Unmarshal([]byte(xm), &rss); err != nil {
		panic(err)
	}

	f, err := os.Open(htmlPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	updates, err := scrapeUpdates(f)
	if err != nil {
		panic(err)
	}

	for _, u := range updates {
		rss.Channel.Items = append(rss.Channel.Items, Item{
			Title:    u.Title,
			Link:     u.Link,
			GUID:     u.Link + "#update",
			Category: "Update",
		})
	}
	out, err := xml.MarshalIndent(rss, "", "  ")
	if err != nil {
		panic(err)
	}

	// Write XML declaration + document
	out = append([]byte(xml.Header), out...)

	if err := os.WriteFile(outputPath, out, 0644); err != nil {
		panic(err)
	}
}
