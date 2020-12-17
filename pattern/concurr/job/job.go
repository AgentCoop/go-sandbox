package job

import (
	"fmt"
	"sync"
)

type JobState int

const (
	New JobState = iota
	WaitingForPrereq
	Running
	Finalizing
	Done
)

type JobTask func(j Job) (func(), func())

type Job interface {
	AddTask(job JobTask) *job
	Run() chan struct{}
	Finish()
	GetRValue() interface{}
	SetRValue(v interface{})
	GetState() JobState
	IsRunning() bool
	IsDone() bool
}

type job struct {
	tasks      	[]func()
	state       JobState
	finishChan 	chan struct{}
	finishWg   	sync.WaitGroup
	prereqWg   	sync.WaitGroup
	finishOnce 	sync.Once
	value      	interface{}
	done		bool
	doneChan	chan struct{}

	mu     		sync.Mutex
	rmu     	sync.RWMutex
	rValue 		interface{} // Access protected by Mutex
	rwValue 	interface{} // Access protected by RWMutex
}

func NewJob(value interface{}) *job {
	j := &job{}
	j.value = value
	j.doneChan = make(chan struct{}, 1)
	return j
}

// A job won't start until all its prerequisites are met
func (j *job) WithPrerequisites(sigs ...<-chan struct{}) *job {
	j.state = WaitingForPrereq
	for _, sig := range sigs {
		j.prereqWg.Add(1)
		go func() {
			for {
				select {
				case <-sig:
					j.prereqWg.Done()
				}
			}
		}()
	}
	return j
}

func (j *job) AddTask(task JobTask) *job {
	j.finishWg.Add(1)
	j.tasks = append(j.tasks, func() {
		run, finish := task(j)
		go func() {
			for {
				select {
				case <-j.finishChan:
					go finish()
					fmt.Println("Done")
					j.finishWg.Done()
					return
				default:
					if j.state == Finalizing {
						// Do nothing and wait for your way own signal
						fmt.Printf("Finalizing")
						continue
					}
					run() // Start another goroutine in order not to block the parent one
				}
			}
		}()
	})
	return j
}

func (j *job) Run() chan struct{} {
	j.state = Running
	j.finishChan = make(chan struct{}, len(j.tasks))
	for _, task := range j.tasks {
		task()
	}
	return j.doneChan
}

func (j *job) finish() {
	j.state = Finalizing
	for i :=0; i < len(j.tasks); i++ {
		j.finishChan <- struct{}{}
	}
	go func(){
		j.finishWg.Wait()
		j.state = Done
		j.doneChan <- struct{}{}
	}()
}

func (j *job) Finish() {
	go func() {
		j.finishOnce.Do(j.finish)
	}()
}

//
// Mutators methods
//
func (j *job) Value() interface{} {
	return j.value
}

func (j *job) GetRValue() interface{} {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.rValue
}

func (j *job) SetRValue(v interface{}) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.rValue = v
}

func (j *job) GetRWValue() interface{} {
	j.rmu.RLock()
	defer j.rmu.RUnlock()
	return j.rwValue
}

func (j *job) SetRWValue(v interface{}) {
	j.rmu.Lock()
	defer j.rmu.Unlock()
	j.rwValue = v
}

func (j *job) GetState() JobState {
	return j.state
}

func (j *job) IsRunning() bool {
	return j.GetState() == Running
}

func (j *job) IsDone() bool {
	return j.GetState() == Done
}
