package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
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
}

func NewJobsGroup() *jobsGroup {
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
	NetFailure chan struct{}
	RandData chan []byte
}

func netFailure() chan struct{} {
	fail := make(chan struct{}, 10000)
	go func() {
		for {
			time.Sleep(time.Second)
			if rand.Intn(10) < 1 {
				fail <- struct{}{}
				//close(fail)
				return
			}
		}
	}()
	return fail
}

func dataSource() chan []byte {
	dataChan := make(chan []byte)
	go func(){
		for {
			data := make([]byte, rand.Intn(6) + 1)
			sleepSecs := rand.Intn(2) + 1
			rand.Read(data)
			time.Sleep(time.Duration(sleepSecs) * time.Second) // Sleep random amount of time
			dataChan <- data
		}
	}()
	return dataChan
}

func upstreamHandler(g *jobsGroup, v interface{}) (func(), func()) {
	return func() {
			u := v.(*UpstreamConn)
			select {
			case <-u.NetFailure:
				fmt.Printf(" ->  Upstream server conn terminated, shutdown\n")
				g.Shutdown()
			case data := <-u.RandData:
				fmt.Printf("Sending data to client [ % X ]\n", data)
				go func() {
					u.Sink <- data
				}()
			case data := <-u.Source:
				fmt.Printf("Client data [ % X ]\n", data)
			default:
			}
	}, func() {
		fmt.Printf("Upstream server shutdown\n")
	}
}

func clientHandler(g *jobsGroup, v interface{}) (func(), func()) {
	return func() {
			u := v.(*UpstreamConn)
			select {
			case <-u.NetFailure:
				fmt.Printf(" -> Client conn terminated, shutdown\n")
				g.Shutdown()
			case data := <-u.RandData:
				fmt.Printf("Sending data to upstream server [ % X ]\n", data)
				go func() {
					u.Source <- data
				}()
			case data := <-u.Sink:
				fmt.Printf("Upstream data [ % X ]\n", data)
			default:
			}
		}, func() {
			fmt.Printf("Client shutdown\n")
		}
}

func main() {
	jg := NewJobsGroup()
	conn := &UpstreamConn{}
	conn.Sink = make(chan []byte)
	conn.Source = make(chan []byte)
	conn.NetFailure = netFailure()
	conn.RandData = dataSource()

	jg.Add(upstreamHandler, conn)
	jg.Add(clientHandler, conn)
	jg.Start()

	time.Sleep(10 * time.Second)
	jg.Shutdown()
}
