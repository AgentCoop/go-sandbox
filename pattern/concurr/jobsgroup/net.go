package main

import (
	"fmt"
	"net"
)

func connRead(conn *net.TCPConn, c chan []byte, hook func([]byte) []byte) error {
	buff := make([]byte, 1024)
	nRead, err := conn.Read(buff)
	if err != nil {
		fmt.Print(err)
		return err
	}
	data := buff[0:nRead]
	if hook != nil {
		data = hook(data)
	}
	c <- data
	return nil
}

func connWrite(conn *net.TCPConn, c chan []byte) error {
	select {
	case data := <- c:
		_, err := conn.Write(data)
		if err != nil {
			fmt.Print(err)
			return err
		}
	}
	return nil
}

func upstreamRead(g *jobsGroup, v interface{}) (func(), func()) {
	return func() {
			u := v.(*UpstreamConn)
			err := connRead(u.UpstreamConn, u.Sink, u.SinkInbound)
			if err != nil {
				g.Shutdown()
			}
		}, func() {
			fmt.Printf("Upstream server shutdown\n")
		}
}

func upstreamWrite(g *jobsGroup, v interface{}) (func(), func()) {
	return func() {
			u := v.(*UpstreamConn)
			err := connWrite(u.UpstreamConn, u.Source)
			if err != nil {
				g.Shutdown()
			}
		}, func() {
			fmt.Printf("Upstream server shutdown\n")
		}
}

func clientRead(g *jobsGroup, v interface{}) (func(), func()) {
	return func() {
			u := v.(*UpstreamConn)
			err := connRead(u.ClientConn, u.Source, u.SourceInbound)
			if err != nil {
				g.Shutdown()
			}
		}, func() {
			//fmt.Printf("Upstream server shutdown\n")
		}
}

func clientWrite(g *jobsGroup, v interface{}) (func(), func()) {
	return func() {
			u := v.(*UpstreamConn)
			err := connWrite(u.UpstreamConn, u.Sink)
			if err != nil {
				g.Shutdown()
			}
		}, func() {
			fmt.Printf("Client close conn\n")
		}
}
