package job_test

import (
	j "github.com/AgentCoop/go-sandbox/pattern/concurr/job"
	"testing"
	"time"
)

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
