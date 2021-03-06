package Parse

import (
	"bytes"

	"golang.org/x/net/html"
)

type Item struct {
	Body   []byte
	BLList []string
	Error  error
}

func (p *Item) parse(n *html.Node) {
	if n.Type == html.ElementNode && n.Data == "a" {
		for i := 0; i < len(n.Attr); i++ {
			if n.Attr[i].Key == "href" {
				p.BLList = append(p.BLList, n.Attr[i].Val)
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		p.parse(c)
	}
}
func (p *Item) Parse() {
	doc, err := html.Parse(bytes.NewReader(p.Body))
	if err != nil {
		p.Error = err
	}
	p.parse(doc)
}

func NewItem(body []byte) *Item {
	return &Item{
		//Id:     id,
		Body:   body,
		BLList: make([]string, 0),
	}
}
