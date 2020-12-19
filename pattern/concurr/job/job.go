package job

import (
	"sync"
	"sync/atomic"
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

type JobTask func(j Job) (func() interface{}, func())

type Job interface {
	AddTask(job JobTask) *TaskInfo
	WithPrerequisites(sigs ...<-chan struct{}) *job
	WithTimeout(duration time.Duration) *job
	WasTimedOut() bool
	Run() chan struct{}
	Cancel()
	Assert(err interface{})
	GetError() interface{}

	GetState() JobState
	// Helper methods to GetState
	IsRunning() bool
	IsDone() bool
	IsCancelled() bool

	GetRValue() interface{}
	SetRValue(v interface{})
	GetRWValue() interface{}
	SetRWValue(v interface{})
}

type TaskInfo struct {
	index 	int
	result 	chan interface{}
	job 	*job
	err 	interface{}
}

func (t *TaskInfo) GetResult() chan interface{} {
	return t.result
}

func (t *TaskInfo) GetJob() Job {
	return t.job
}

type job struct {
	tasks      			[]func()
	cancelTasks      	[]func()
	runningTasks 		int
	runningTaskMu		sync.Mutex
	failedTasksCounter	int32
	runningTasksCounter	int32
	state      			JobState
	timedoutFlag		bool
	timeout				time.Duration

	errorChan			chan interface{}
	errorInfo			interface{}
	doneChan    		chan struct{}
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

func (j *job) Assert(err interface{}) {
	if err != nil {
		j.errorChan <- err
		// Now time to panic to stop normal goroutine execution from which Assert method was called.
		panic(err)
	}
}

func (j *job) AddTask(task JobTask) *TaskInfo {
	taskInfo := &TaskInfo{}
	taskInfo.index = len(j.tasks)
	taskInfo.result =  make(chan interface{}, 1)
	taskBody := func() {
		run, cancel := task(j)
		j.cancelTasks = append(j.cancelTasks, cancel)
		go func() {
			defer func() {
				j.runningTaskMu.Lock()
				j.runningTasksCounter--
				j.runningTaskMu.Unlock()

				if r := recover(); r != nil {
					atomic.AddInt32(&j.failedTasksCounter, 1)
				}
			}()

			for {
				result := run()
				j.stateMu.Lock()
				switch {
				case j.state != Running:
					j.stateMu.Unlock()
					taskInfo.result <- result
					return
				default:
					j.stateMu.Unlock()
				}

				if result != nil {
					taskInfo.result <- result
					return
				}
			}
		}()
	}
	j.tasks = append(j.tasks, taskBody)
	return taskInfo
}

// Concurrently executes all tasks in the job.
func (j *job) Run() chan struct{} {
	nTasks := len(j.tasks)
	j.errorChan = make(chan interface{}, nTasks)
	j.runningTasksCounter = int32(nTasks)
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

	// Error listener
	go func() {
		for {
			select {
			case err := <- j.errorChan:
				j.errorInfo = err
				j.Cancel()
				return
			}
		}
	}()

	// Diptaches Done signal sif there is no
	go func() {
		for {
			j.runningTaskMu.Lock()
			if j.runningTasksCounter == 0 {
				j.state = Done
				j.runningTaskMu.Unlock()
				j.doneChan <- struct{}{}
				return
			}
			j.runningTaskMu.Unlock()
		}
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

	if (j.timedoutFlag) {
		j.state = Cancelled
		j.doneChan <- struct{}{}
	}

	//j.doneTasksWg.Done()
}

func (j *job) WasTimedOut() bool {
	return j.timedoutFlag
}

func (j *job) GetError() interface{} {
	return j.errorInfo
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
