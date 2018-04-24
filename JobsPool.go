package main

import (
	"fmt"
	"sync"

	"github.com/LepikovStan/bCrawler/Queue"
	"github.com/LepikovStan/bCrawler/Workers"
)

type JobsPool struct {
	out                    chan *Workers.CrawlJob
	Queue                  *Queue.Q
	ErrorQueue             *Queue.Q
	mu                     *sync.RWMutex
	jobsCount, errorsCount int
}

func (j *JobsPool) IncErrCount(num int) {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.errorsCount += num
}
func (j *JobsPool) DecErrCount(num int) {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.errorsCount -= num
}

func (j *JobsPool) incCount() {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.jobsCount++
}
func (j *JobsPool) decCount() {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.jobsCount--
}
func (j JobsPool) Empty() bool {
	return j.Queue.Len() == 0 && j.jobsCount <= 0 && j.errorsCount <= 0
}
func (j *JobsPool) RunJob() {
	item := j.Queue.Pop()
	if item == nil {
		return
	}

	Job := item.(*Workers.CrawlJob)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				if fmt.Sprintf("%s", r) != "send on closed channel" {
					panic(r)
				}
			}
		}()
		j.out <- Job
	}()
	j.incCount()
}
func (j *JobsPool) DoneJob() {
	if j.jobsCount > 0 {
		j.decCount()
	}
}
func NewJobsPool(out chan *Workers.CrawlJob) *JobsPool {
	return &JobsPool{
		out:        out,
		Queue:      Queue.NewQ(),
		ErrorQueue: Queue.NewQ(),
		mu:         &sync.RWMutex{},
	}
}
