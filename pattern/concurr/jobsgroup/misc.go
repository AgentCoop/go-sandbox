package main

import (
	"math/rand"
	"time"
)

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
