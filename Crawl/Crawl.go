package Crawl

import (
	"io/ioutil"
	"net/http"
)

type Item struct {
	//Id    int
	Url   string
	Body  []byte
	Error error
}

func (c *Item) Crawl() error {
	resp, err := http.Get(c.Url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	c.Body = body
	return nil
}

func NewItem(url string) *Item {
	return &Item{
		Url: url,
	}
}
