package Workers

import (
	"fmt"

	"sync"

	"github.com/LepikovStan/bCrawler/Crawl"
)

type Crawler struct {
	Id      int
	In, Out chan *CrawlJob
	Wg      *sync.WaitGroup
}

type CrawlJob struct {
	Id, Depth int
	CrawlItem *Crawl.Item
}

func (c Crawler) Run() {
	fmt.Println("Run Crawler", c.Id)
	for Job := range c.In {
		Job.CrawlItem.Crawl()
		c.Out <- Job
	}
	c.Wg.Done()
}

func (c Crawler) Shutdown() {
	c.Wg.Done()
}

func NewCrawlJob(Id, Depth int, CrawlItem *Crawl.Item) *CrawlJob {
	return &CrawlJob{
		Id:        Id,
		Depth:     Depth,
		CrawlItem: CrawlItem,
	}
}

func NewCrawl(Id int, In, Out chan *CrawlJob, wg *sync.WaitGroup) *Crawler {
	return &Crawler{
		Id:  Id,
		In:  In,
		Out: Out,
		Wg:  wg,
	}
}
