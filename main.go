package main

import (
	"fmt"
	"time"

	"sync"

	"os"
	"os/signal"

	"github.com/LepikovStan/bCrawler/Crawl"
	"github.com/LepikovStan/bCrawler/Parse"
	"github.com/LepikovStan/bCrawler/Workers"
)

const (
	CrawlersCount = 5
	ParsersCount  = 5
	MaxDepth      = 3
	RetryTime     = time.Second * 3
)

var (
	JobId = 0
	Count = 0
	cPool = make([]*Workers.Crawler, CrawlersCount)
	pPool = make([]*Workers.Parser, ParsersCount)
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

func ErrorsDispatcher(Out chan *Workers.CrawlJob, JobsList *JobsPool) {
	for {
		time.Sleep(RetryTime)
		Job := JobsList.ErrorQueue.Pop().(*Workers.CrawlJob)
		if Job == nil {
			continue
		}
		Job.Retry--
		go func() {
			defer func() {
				if r := recover(); r != nil {
					if fmt.Sprintf("%s", r) != "send on closed channel" {
						panic(r)
					}
				}
			}()
			Out <- Job
		}()
	}
}

func Dispatcher(CrawlIn, CrawlOut chan *Workers.CrawlJob, ParseIn, ParseOut chan *Workers.ParseJob, JobsList *JobsPool) {
	for {
		select {
		case Job := <-CrawlOut:
			if Job == nil {
				continue
			}
			if Job.CrawlItem.Error != nil {
				fmt.Println("Crawled with error", Job.Id, Job.CrawlItem.Url, len(Job.CrawlItem.Body), Job.CrawlItem.Error)
				JobsList.DoneJob()

				if Job.Retry == Workers.Retry {
					JobsList.IncErrCount(Workers.Retry)
				} else {
					JobsList.DecErrCount(1)
				}

				if JobsList.Empty() && Job.Depth+1 > MaxDepth {
					Shutdown(CrawlIn)
					return
				}
				JobsList.ErrorQueue.Push(Job)
				for i := 0; i < CrawlersCount-len(CrawlIn); i++ {
					JobsList.RunJob()
				}
				//JobsList.RunJob()
				continue
			}
			fmt.Println("Crawled", Job.Id, Job.CrawlItem.Url, len(Job.CrawlItem.Body), Job.CrawlItem.Error)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						if fmt.Sprintf("%s", r) != "send on closed channel" {
							panic(r)
						}
					}
				}()
				ParseIn <- Workers.NewParseJob(Job.Id, Job.Depth, Parse.NewItem(Job.CrawlItem.Body))
			}()
		case Job := <-ParseOut:
			if Job == nil {
				continue
			}
			JobsList.DoneJob()
			Count++
			fmt.Println("Parsed", Job.Id, len(Job.ParseItem.BLList), Job.ParseItem.Error)

			if JobsList.Empty() && Job.Depth+1 > MaxDepth {
				Shutdown(CrawlIn)
				return
			}
			if Job.ParseItem.Error != nil {
				JobsList.ErrorQueue.Push(Job)
				for i := 0; i < CrawlersCount-len(CrawlIn); i++ {
					JobsList.RunJob()
				}
				//JobsList.RunJob()
				continue
			}
			if Job.Depth+1 <= MaxDepth {
				for i := 0; i < len(Job.ParseItem.BLList); i++ {
					Job := Workers.NewCrawlJob(NewJobId(), Job.Depth+1, Crawl.NewItem(Job.ParseItem.BLList[i]))
					JobsList.Queue.Push(Job)
				}
			}
			for i := 0; i < CrawlersCount-len(CrawlIn); i++ {
				JobsList.RunJob()
			}
			//JobsList.RunJob()
		}
	}
}

func Shutdown(CrawlIn chan *Workers.CrawlJob) {
	close(CrawlIn)
}

func interceptInterrupt(CrawlIn chan *Workers.CrawlJob) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c
	fmt.Println()
	fmt.Println("Force Shutdown...", len(CrawlIn))
	Shutdown(CrawlIn)
}

func main() {
	s := time.Now()

	CrawlIn := make(chan *Workers.CrawlJob, CrawlersCount)
	CrawlOut := make(chan *Workers.CrawlJob, CrawlersCount)
	ParseIn := make(chan *Workers.ParseJob, ParsersCount)
	ParseOut := make(chan *Workers.ParseJob, ParsersCount)
	//StartList := []string{"http://ya.ru", "http://go-search.org"}
	//StartList := []string{"ya.ru", "http://ya.ru"}
	StartList := []string{"http://ya.ru"}
	JobsList := NewJobsPool(CrawlIn)

	cwg := &sync.WaitGroup{}
	pwg := &sync.WaitGroup{}

	RunCrawlers(CrawlIn, CrawlOut, cwg)
	RunParsers(ParseIn, ParseOut, pwg)
	go Dispatcher(CrawlIn, CrawlOut, ParseIn, ParseOut, JobsList)
	go ErrorsDispatcher(CrawlIn, JobsList)
	go interceptInterrupt(CrawlIn)

	for i := 0; i < len(StartList); i++ {
		Job := Workers.NewCrawlJob(NewJobId(), 0, Crawl.NewItem(StartList[i]))
		JobsList.Queue.Push(Job)
	}
	for i := 0; i < CrawlersCount; i++ {
		JobsList.RunJob()
	}

	cwg.Wait()
	close(CrawlOut)

	close(ParseIn)
	pwg.Wait()
	close(ParseOut)

	fmt.Println("Done...", Count)
	fmt.Println(time.Now().Sub(s))
}
