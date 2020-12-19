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
	return [...]string{"New", "WaitingForPrereq", "Running", "Cancelling", "Cancelled","Done"}[s]
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
}

type job struct {
	tasks      			[]func()
	cancelTasks      	[]func()
	// If set to true stop execution of all tasks
	stopExec			bool
	runningTasks 		int
	runningTaskMu		sync.Mutex
	state      			JobState
	timedoutFlag		bool
	cancelFlag			bool
	timeoutChan			chan time.Time
	timeout				time.Duration

	cancelChan  chan struct{}
	doneChan    chan struct{}
	cancelOnce  sync.Once
	doneTasksWg sync.WaitGroup
	prereqWg    sync.WaitGroup

	value      			interface{}
	rValue 				interface{} // Access protected by Mutex
	rwValue 			interface{} // Access protected by RWMutex

	mu     				sync.Mutex
	rmu     			sync.RWMutex
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
				j.mu.Lock()

				switch {
				case j.state != Running:
					j.mu.Unlock()
					j.incRunningTasksCounter(-1)
					return
				default:
					j.mu.Unlock()
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
	j.cancelChan = make(chan struct{}, len(j.tasks))

	// Start timer that will cancel and mark the job as timed out if needed
	if j.timeout > 0 {
		go func() {
			ch := time.After(j.timeout)
			for {
				select {
				case <-ch:
					j.timedoutFlag = true
					j.Cancel()
					return
				}
			}
		}()
	}

	// Set final state and send Done signal
	go func() {
		j.doneTasksWg.Wait()
		j.mu.Lock()
		defer j.mu.Unlock()
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

func (j *job) cancel() {
	for i :=0; i < len(j.tasks); i++ {
		j.cancelChan <- struct{}{}
	}
}

func (j *job) Cancel() {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.state != Running { return }
	j.state = Cancelling

	for _, cancel := range j.cancelTasks {
		go cancel()
	}

	for i := 0; i < j.runningTasks; i++ {
		j.doneTasksWg.Done()
	}
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
	return j.state == Running
}

func (j *job) IsCancelled() bool {
	return j.state == Cancelled || j.state == Cancelling
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
