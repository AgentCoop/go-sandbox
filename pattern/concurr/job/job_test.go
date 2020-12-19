package job_test

import (
	j "github.com/AgentCoop/go-sandbox/pattern/concurr/job"
	"sync"
	"testing"
	"time"
)

var counter int
var mu sync.Mutex

func t(info *j.TaskInfo) {
	info.GetResult()
}

func incCounterJob(j j.Job) (func() interface{}, func()) {
	return func() interface{} {
		mu.Lock()
		defer mu.Unlock()
		counter++
		return counter
	}, func() { }
}

func squareJob(num int) j.JobTask {
	return func(j j.Job) (func() interface{}, func()) {
		return func() interface{} {
			return num * num
		}, func() { }
	}
}

func sleepIncCounterJob(sleep time.Duration) j.JobTask {
	return func(j j.Job) (func() interface{}, func()) {
		return func() interface{} {
			if sleep > 0 {
				time.Sleep(sleep)
			}
			mu.Lock()
			defer mu.Unlock()
			counter++
			return counter
		}, func() { }
	}
}

func signalAfter(t time.Duration, fn func()) chan struct{} {
	ch := make(chan struct{})
	go func() {
		time.Sleep(t)
		if fn != nil {
			fn()
		}
		ch <- struct{}{}
	}()
	return ch
}

func TestFinish(T *testing.T) {
	job := j.NewJob(nil)
	job.AddTask(func(j j.Job) (func() interface{}, func()) {
		return func() interface{} {
			time.Sleep(10 * time.Millisecond)
			j.SetRValue(1)
			j.Cancel()
			return false
		}, func() {
			if j.GetRValue() != 1 {
				T.Fatalf("got %d, expected %d\n", j.GetRValue(), 1)
			}
		}
	})
	job.AddTask(func(j j.Job) (func() interface{}, func()) {
		return func() interface{} {
			time.Sleep(30 * time.Millisecond)
			if ! j.IsRunning() { return false }
			j.SetRValue(2)
			return false
		}, func() {
			if j.GetRValue() != 1 {
				T.Fatalf("got %d, expected %d\n", j.GetRValue(), 1)
			}
		}
	})

	<-job.Run()

	if ! job.IsCancelled() {
		T.Fatalf("got state %v, expected %v\n", job.GetState(), j.Cancelled)
	}
}

func TestPrereq(T *testing.T) {
	var counter int
	p1 := signalAfter(10 * time.Millisecond, func() { counter++ })
	p2 := signalAfter(20 * time.Millisecond, func() { counter++ })
	job := j.NewJob(nil)
	job.WithPrerequisites(p1, p2)
	job.AddTask(func(j j.Job) (func() interface{}, func()) {
		return func() interface{} {
				if counter != 2 {
					T.Fatalf("got %d, expected %d\n", counter, 2)
				}
				j.Cancel()
				return false
			}, func() {

			}
	})
	<-job.Run()
}

func TestDone(T *testing.T) {
	counter = 0
	job := j.NewJob(nil)
	job.AddTask(incCounterJob)
	job.AddTask(incCounterJob)
	<-job.Run()
	if ! job.IsDone() || counter != 2 {
		T.Fail()
	}
}

func TestTimeout(T *testing.T) {
	// Must succeed
	counter = 0
	job := j.NewJob(nil).WithTimeout(120 * time.Millisecond)
	for i := 0; i < 100; i++ {
		job.AddTask(sleepIncCounterJob(time.Duration(i + 1) * time.Millisecond))
	}
	<-job.Run()
	if ! job.IsDone() || counter != 100 {
		T.Fatalf("expected counter 100, got %d\n", counter)
	}
	// Must be cancelled
	counter = 0
	job = j.NewJob(nil)
	job.WithTimeout(15 * time.Millisecond)
	job.AddTask(sleepIncCounterJob(10 * time.Millisecond))
	job.AddTask(sleepIncCounterJob(99999 * time.Second)) // Must not block run method
	<-job.Run()
	if ! job.IsCancelled() || counter != 1 {
		T.Fail()
	}
}

func TestTaskResult(T *testing.T) {
	// Must succeed
	counter = 0
	job := j.NewJob(nil).WithTimeout(10 * time.Millisecond)
	task1 := job.AddTask(squareJob(3))
	task2 := job.AddTask(sleepIncCounterJob(15 * time.Millisecond))
	<-job.Run()
	if ! job.IsCancelled() || counter != 0 {
		T.Fatalf("expected: counter 0, state Done; got: %d %s\n", counter, job.GetState())
	}
	select {
	case num := <- task1.GetResult():
		if num != 9 { T.Fatalf("expected: 0; got: %d\n", num) }
	case <- task2.GetResult():
		T.Fatal()
	default:
		T.Fatal()
	}
}