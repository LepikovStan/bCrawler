package main

import (
	"fmt"
	"time"

	"sync"

	"github.com/LepikovStan/bCrawler/Crawl"
	"github.com/LepikovStan/bCrawler/Parse"
	"github.com/LepikovStan/bCrawler/Queue"
	"github.com/LepikovStan/bCrawler/Workers"
)

const (
	CrawlersCount = 10
	ParsersCount  = 5
	MaxDepth      = 2
)

var (
	JobId     = 0
	JobsCount = 0
	Count     = 0
	cPool     = make([]*Workers.Crawler, CrawlersCount)
	pPool     = make([]*Workers.Parser, ParsersCount)
)

func NewJobId() int {
	JobId++
	return JobId
}

func RunCrawlers(CrawlIn, CrawlOut chan *Workers.CrawlJob, wg *sync.WaitGroup) {
	for i := 0; i < CrawlersCount; i++ {
		wg.Add(1)
		cWorker := Workers.NewCrawl(i, CrawlIn, CrawlOut, wg)
		cPool[i] = cWorker
		go cWorker.Run()
	}
}
func RunParsers(ParseIn, ParseOut chan *Workers.ParseJob, wg *sync.WaitGroup) {
	for i := 0; i < ParsersCount; i++ {
		wg.Add(1)
		pWorker := Workers.NewParse(i, ParseIn, ParseOut, wg)
		pPool[i] = pWorker
		go pWorker.Run()
	}
}

func Dispatcher(CrawlIn, CrawlOut chan *Workers.CrawlJob, ParseIn, ParseOut chan *Workers.ParseJob, JobsList *JobsPool) {
	for {
		select {
		case Job := <-CrawlOut:
			fmt.Println("Crawled", Job.Id, len(Job.CrawlItem.Body), JobsList.Queue.Len(), JobsList.Count())
			if Job.CrawlItem.Error == nil {
				go func() {
					ParseIn <- Workers.NewParseJob(Job.Id, Job.Depth, Parse.NewItem(Job.CrawlItem.Body))
				}()
			}
		case Job := <-ParseOut:
			JobsList.DecCount()
			Count++
			fmt.Println("Parsed", Job.Id, len(Job.ParseItem.BLList), JobsList.Queue.Len(), JobsList.Count())

			if JobsList.Queue.Len() == 0 && Job.Depth+1 > MaxDepth && JobsList.Count() == 0 {
				ShutdownWorkers(CrawlIn, CrawlOut, ParseIn, ParseOut)
				return
			}
			if Job.Depth+1 <= MaxDepth {
				for i := 0; i < len(Job.ParseItem.BLList); i++ {
					Job := Workers.NewCrawlJob(NewJobId(), Job.Depth+1, Crawl.NewItem(Job.ParseItem.BLList[i]))
					JobsList.Queue.Push(Job)
				}
			}

			JobsList.RunJob()

		}
	}
}

func ShutdownWorkers(CrawlIn, CrawlOut chan *Workers.CrawlJob, ParseIn, ParseOut chan *Workers.ParseJob) {
	close(CrawlIn)
	close(CrawlOut)
	close(ParseIn)
	close(ParseOut)
	fmt.Println("Done...", Count)
}

type JobsPool struct {
	out       chan *Workers.CrawlJob
	Queue     *Queue.Q
	mu        *sync.RWMutex
	JobsCount int
}

func (j *JobsPool) IncCount() {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.JobsCount++
}
func (j *JobsPool) DecCount() {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.JobsCount--
}
func (j JobsPool) Count() int {
	j.mu.RLock()
	defer j.mu.RUnlock()

	return j.JobsCount
}
func (j *JobsPool) RunJob() {
	item := j.Queue.Pop()
	if item == nil {
		return
	}
	Job := item.(*Workers.CrawlJob)
	go func() {
		j.out <- Job
	}()
	j.IncCount()
}
func NewJobs(out chan *Workers.CrawlJob, queue *Queue.Q) *JobsPool {
	return &JobsPool{
		out:   out,
		Queue: queue,
		mu:    &sync.RWMutex{},
	}
}

func main() {
	s := time.Now()

	CrawlIn := make(chan *Workers.CrawlJob)
	CrawlOut := make(chan *Workers.CrawlJob)
	ParseIn := make(chan *Workers.ParseJob)
	ParseOut := make(chan *Workers.ParseJob)
	//StartList := []string{"http://yandex.ru", "http://ya.ru", "http://golang.org"}
	StartList := []string{"http://ya.ru"}
	queue := Queue.NewQ()
	JobsList := NewJobs(CrawlIn, queue)

	cwg := &sync.WaitGroup{}
	pwg := &sync.WaitGroup{}

	RunCrawlers(CrawlIn, CrawlOut, cwg)
	RunParsers(ParseIn, ParseOut, pwg)
	go Dispatcher(CrawlIn, CrawlOut, ParseIn, ParseOut, JobsList)

	for i := 0; i < len(StartList); i++ {
		Job := Workers.NewCrawlJob(NewJobId(), 0, Crawl.NewItem(StartList[i]))
		queue.Push(Job)
	}
	for i := 0; i < CrawlersCount; i++ {
		JobsList.RunJob()
	}

	cwg.Wait()
	//close(CrawlIn)

	pwg.Wait()
	//close(ParseIn)
	//w := new(Worker)

	//crwlrJob := Crawler.NewItem("http://ya.ru")
	//err := crwlrJob.Crawl()
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println("crawled", time.Now().Sub(s))
	//
	//s1 := time.Now()
	//prsrJob := Parser.NewItem(crwlrJob.Body)
	//prsrJob.Parse()
	//fmt.Println("parsed", time.Now().Sub(s1))
	//
	//fmt.Println(len(prsrJob.BLList))

	fmt.Println(time.Now().Sub(s))
}
