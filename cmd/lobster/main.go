package main

import (
	"fmt"
	"io"
	"net/http"
)

const feedURL = "https://lobste.rs/top/rss"

func main() {
	resp, err := http.Get(feedURL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(resp.Status)
	}

	rss, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Print(string(rss))
}
