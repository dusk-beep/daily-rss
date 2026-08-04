package main

import (
	"encoding/json"
	"fmt"
	"html"
	"os"
	"strings"

	xhtml "golang.org/x/net/html"
)

type Response struct {
	Solution struct {
		Response string `json:"response"`
	} `json:"solution"`
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

	if pre == nil || pre.FirstChild == nil {
		return "", fmt.Errorf("no <pre> found")
	}

	return html.UnescapeString(pre.FirstChild.Data), nil
}

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	var r Response
	if err := json.Unmarshal(data, &r); err != nil {
		panic(err)
	}

	xml, err := extractXML(r.Solution.Response)
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile("feed.xml", []byte(xml), 0644); err != nil {
		panic(err)
	}
}
