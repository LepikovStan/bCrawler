package Workers

import (
	"fmt"
	"sync"

	"github.com/LepikovStan/bCrawler/Parse"
)

type Parser struct {
	Id      int
	In, Out chan *ParseJob
	Wg      *sync.WaitGroup
}

type ParseJob struct {
	Id, Depth int
	ParseItem *Parse.Item
}

func (p Parser) Run() {
	fmt.Println("Run Parser", p.Id)
	for Job := range p.In {
		Job.ParseItem.Parse()
		p.Out <- Job
	}
	p.Wg.Done()
}

func (p Parser) Shutdown() {
	p.Wg.Done()
}

func NewParseJob(Id, Depth int, ParseItem *Parse.Item) *ParseJob {
	return &ParseJob{
		Id:        Id,
		Depth:     Depth,
		ParseItem: ParseItem,
	}
}

func NewParse(Id int, In, Out chan *ParseJob, wg *sync.WaitGroup) *Parser {
	return &Parser{
		Id:  Id,
		In:  In,
		Out: Out,
		Wg:  wg,
	}
}
