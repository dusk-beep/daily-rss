package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
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
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	var r Response
	if err := json.Unmarshal(data, &r); err != nil {
		panic(err)
	}

	xm, err := extractXML(r.Solution.Response)
	if err != nil {
		panic(err)
	}

	dec := xml.NewDecoder(strings.NewReader(xm))
	for {
		if _, err := dec.Token(); err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
	}

	if err := os.WriteFile("feed.xml", []byte(xm), 0644); err != nil {
		panic(err)
	}
}
