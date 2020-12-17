package job_test

import (
	j "github.com/AgentCoop/go-sandbox/pattern/concurr/job"
	"testing"
	"time"
)

//func p(msg string, a ...interface{}) {
//	fmt.Printf(msg, a)
//}

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
	job.AddTask(func(j j.Job) (func(), func()) {
		return func() {
			time.Sleep(10 * time.Millisecond)
			j.SetRValue(1)
			j.Finish()
		}, func() {
			if j.GetRValue() != 1 {
				T.Fatalf("got %d, expected %d\n", j.GetRValue(), 1)
			}
		}
	})
	job.AddTask(func(j j.Job) (func(), func()) {
		return func() {
			time.Sleep(30 * time.Millisecond)
			if ! j.IsRunning() { return }
			j.SetRValue(2)
		}, func() {
			if j.GetRValue() != 1 {
				T.Fatalf("got %d, expected %d\n", j.GetRValue(), 1)
			}
		}
	})
	<-job.Run()
}

func TestPrereq(T *testing.T) {
	var counter int
	p1 := signalAfter(10 * time.Millisecond, func() { counter++ })
	p2 := signalAfter(20 * time.Millisecond, func() { counter++ })
	job := j.NewJob(nil)
	job.WithPrerequisites(p1, p2)
	job.AddTask(func(j j.Job) (func(), func()) {
		return func() {
				if counter != 2 {
					T.Fatalf("got %d, expected %d\n", counter, 2)
				}
				j.Finish()
			}, func() {

			}
	})
	<-job.Run()
}