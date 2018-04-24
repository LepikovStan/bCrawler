package Crawl

import (
	"io/ioutil"
	"net/http"
)

type Item struct {
	Url   string
	Body  []byte
	Error error
}

func (c *Item) Crawl() {
	resp, err := http.Get(c.Url)
	if err != nil {
		c.Error = err
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.Error = err
		return
	}
	c.Body = body
}

func NewItem(url string) *Item {
	return &Item{
		Url: url,
	}
}
