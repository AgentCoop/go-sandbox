package main

import (
	"fmt"
	"net"
	"sync"
)

type JobFunc func(bool, interface{}) bool

type JobsGroup interface {
	Add(job func(g *jobsGroup, v interface{}) bool, val interface{})
	Start()
	Shutdown()
}

type jobsGroup struct {
	jobs         []func()
	shutdownChan chan struct{}
	wg           sync.WaitGroup
	shutdownOnce sync.Once
	SafeValue		interface{}
}

func NewJobsGroup(v interface{}) *jobsGroup {
	g := &jobsGroup{}
	g.shutdownChan = make(chan struct{}, 1)
	return g
}

type F func()

func (g *jobsGroup) Add(job func(g *jobsGroup, v interface{}) (func(), func()), v interface{}) *jobsGroup {
	g.wg.Add(1)
	g.jobs = append(g.jobs, func() {
		go func() {
			run, cancel := job(g, v)
			for {
				select {
				case <-g.shutdownChan:
					cancel()
					g.wg.Done()
					return
				default:
					run() // Start another goroutine in order not to block the parent one
				}
			}
		}()
	})
	return g
}

func (g *jobsGroup) shutdown() {
	for i :=0; i < len(g.jobs); i++ {
		g.shutdownChan <- struct{}{}
	}
	g.wg.Wait()
}

func (g *jobsGroup) Shutdown() *jobsGroup {
	go func() {
		g.shutdownOnce.Do(g.shutdown)
	}()
	return g
}

func (g *jobsGroup) Start() *jobsGroup {
	for i, job := range g.jobs {
		fmt.Printf("Starting job #%d\n", i + 1)
		go job()
	}
	return g
}

//
// Example of usage
//

type UpstreamConn struct {
	Source chan []byte
	Sink chan []byte

	SourceInbound func(data []byte) []byte
	SourceOutbound func(data []byte) []byte

	SinkInbound func(data []byte) []byte
	SinkOutbound func(data []byte) []byte

	UpstreamConn *net.TCPConn
	ClientConn *net.TCPConn

	NetFailure chan struct{}
	RandData chan []byte
}
