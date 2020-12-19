package job

import (
	"sync"
	"time"
)

type JobState int

const (
	New JobState = iota
	WaitingForPrereq
	Running
	Cancelling
	Cancelled
	Done
)

func (s JobState) String() string {
	return [...]string{"New", "WaitingForPrereq", "Running", "Cancelling", "Cancelled", "Done"}[s]
}

type JobTask func(j Job) (func() bool, func())

type Job interface {
	AddTask(job JobTask) *job
	WithPrerequisites(sigs ...<-chan struct{}) *job
	WithTimeout(duration time.Duration) *job
	Run() chan struct{}
	Cancel()
	GetState() JobState
	GetRValue() interface{}
	SetRValue(v interface{})
	IsRunning() bool
	IsDone() bool
	IsCancelled() bool
	WasTimedOut() bool
}

type job struct {
	tasks      			[]func()
	cancelTasks      	[]func()
	runningTasks 		int
	runningTaskMu		sync.Mutex
	state      			JobState
	timedoutFlag		bool
	timeout				time.Duration

	doneChan    		chan struct{}
	doneTasksWg 		sync.WaitGroup
	prereqWg    		sync.WaitGroup

	value      			interface{}
	rValue 				interface{}
	rValMu				sync.Mutex
	rwValue 			interface{}
	rwValMu				sync.RWMutex

	stateMu 			sync.Mutex
}

func NewJob(value interface{}) *job {
	j := &job{}
	j.state = New
	j.value = value
	j.doneChan = make(chan struct{}, 1)
	return j
}

// A job won't start until all its prerequisites are met
func (j *job) WithPrerequisites(sigs ...<-chan struct{}) *job {
	j.state = WaitingForPrereq
	j.prereqWg.Add(len(sigs))
	for _, sig := range sigs {
		s := sig // Binds loop variable to go closure
		go func() {
			for {
				select {
				case <-s:
					j.prereqWg.Done()
					return
				}
			}
		}()
	}
	return j
}

func (j *job) WithTimeout(t time.Duration) *job {
	j.timeout = t
	return j
}

func (j *job) AddTask(task JobTask) *job {
	j.doneTasksWg.Add(1)
	j.tasks = append(j.tasks, func() {
		run, cancel := task(j)
		j.cancelTasks = append(j.cancelTasks, cancel)
		go func() {
			j.incRunningTasksCounter(1)
			for {
				done := run()
				j.stateMu.Lock()

				switch {
				case j.state != Running:
					j.stateMu.Unlock()
					j.incRunningTasksCounter(-1)
					return
				default:
					j.stateMu.Unlock()
				}

				if done {
					j.doneTasksWg.Done()
					j.incRunningTasksCounter(-1)
					return
				}
			}
		}()
	})
	return j
}

func (j *job) Run() chan struct{} {
	// Start timer that will cancel and mark the job as timed out if needed
	if j.timeout > 0 {
		go func() {
			ch := time.After(j.timeout)
			for {
				select {
				case <-ch:
					if j.state == Running {
						j.timedoutFlag = true
						j.Cancel()
					}
					return
				}
			}
		}()
	}

	// Sets final state and dispatches Done signal
	go func() {
		j.doneTasksWg.Wait()
		j.stateMu.Lock()
		defer j.stateMu.Unlock()
		switch j.state {
		case Cancelling:
			j.state = Cancelled
		case Running:
			j.state = Done
		}
		j.doneChan <- struct{}{}
	}()

	if j.state == WaitingForPrereq {
		j.prereqWg.Wait()
	}

	j.state = Running
	for _, task := range j.tasks {
		task()
	}
	return j.doneChan
}

func (j *job) Cancel() {
	j.stateMu.Lock()
	defer j.stateMu.Unlock()

	if j.state != Running { return }
	j.state = Cancelling

	for _, cancel := range j.cancelTasks {
		go cancel()
	}

	for i := 0; i < j.runningTasks; i++ {
		j.doneTasksWg.Done()
	}
}

func (j *job) WasTimedOut() bool {
	return j.timedoutFlag
}

//
// Mutators methods
//
func (j *job) Value() interface{} {
	return j.value
}

func (j *job) GetRValue() interface{} {
	j.rValMu.Lock()
	defer j.rValMu.Unlock()
	return j.rValue
}

func (j *job) SetRValue(v interface{}) {
	j.rValMu.Lock()
	defer j.rValMu.Unlock()
	j.rValue = v
}

func (j *job) GetRWValue() interface{} {
	j.rwValMu.RLock()
	defer j.rwValMu.RUnlock()
	return j.rwValue
}

func (j *job) SetRWValue(v interface{}) {
	j.rwValMu.Lock()
	defer j.rwValMu.Unlock()
	j.rwValue = v
}

func (j *job) GetState() JobState {
	return j.state
}

func (j *job) IsRunning() bool {
	return j.state == Running
}

func (j *job) IsCancelled() bool {
	return j.state == Cancelled
}

func (j *job) IsDone() bool {
	return j.state == Done
}

//
// Helper methods
//
func (j *job) incRunningTasksCounter(inc int) {
	j.runningTaskMu.Lock()
	defer j.runningTaskMu.Unlock()
	j.runningTasks += inc
}
