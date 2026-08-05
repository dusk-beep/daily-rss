package main

import (
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Update struct {
	Title string
	Link  string
}

func scrapeUpdates(r io.Reader) ([]Update, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	var updates []Update

	doc.Find("#text-6 .textwidget p a").Each(func(i int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok {
			return
		}

		title := strings.TrimSpace(s.Text())
		if title == "" {
			return
		}

		updates = append(updates, Update{
			Title: title,
			Link:  href,
		})
	})

	return updates, nil
}
