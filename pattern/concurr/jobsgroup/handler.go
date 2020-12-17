package main

import "fmt"

func upstreamHandler(g *jobsGroup, v interface{}) (func(), func()) {
	return func() {
			u := v.(*UpstreamConn)
			select {
			//case <-u.NetFailure:
			//	fmt.Printf(" ->  Upstream server conn terminated, shutdown\n")
			//	g.Shutdown()
			//case data := <-u.RandData:
			//	fmt.Printf("Sending data to client [ % X ]\n", data)
			//	go func() {
			//		u.Sink <- data
			//	}()
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
			//case <-u.NetFailure:
			//	fmt.Printf(" -> Client conn terminated, shutdown\n")
			//	g.Shutdown()
			//case data := <-u.RandData:
			//	fmt.Printf("Sending data to upstream server [ % X ]\n", data)
			//	go func() {
			//		u.Source <- data
			//	}()
			case data := <-u.Sink:
				fmt.Printf("Upstream data [ % X ]\n", data)
			default:
			}
		}, func() {
			fmt.Printf("Client shutdown\n")
		}
}
