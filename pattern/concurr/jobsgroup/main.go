package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

func setUpConn() (*net.TCPConn, *net.TCPConn) {
	service := ":8000"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	plis, err := net.ListenTCP("tcp4", tcpAddr)
	if err != nil {
		panic(err)
	}
	pconn, acceptErr := plis.AcceptTCP()
	if acceptErr != nil {
		panic(acceptErr)
	}
	upStreamAddr, err := net.ResolveTCPAddr("tcp4", ":8001")
	upstreamConn, dialErr := net.DialTCP("tcp4", nil, upStreamAddr)
	if dialErr != nil {
		panic(dialErr)
	}
	return pconn, upstreamConn
}

func main() {

	p, upstream := setUpConn()

	jg := NewJobsGroup()
	conn := &UpstreamConn{}
	conn.Sink = make(chan []byte)
	conn.Source = make(chan []byte)

	conn.SinkInbound = func(data []byte) []byte {
		fmt.Printf("Client data [ % X ]\n", data)
		return data
	}

	conn.SourceInbound = func(data []byte) []byte {
		fmt.Printf("Upstream data [ % X ]\n", data)
		s := string(data)
		upperCase := strings.ToUpper(s)
		return []byte(upperCase)
	}

	//conn.NetFailure = netFailure()
	//conn.RandData = dataSource()

	conn.ClientConn = p
	conn.UpstreamConn = upstream

	//jg.Add(upstreamHandler, conn)
	//jg.Add(clientHandler, conn)

	jg.Add(upstreamRead,conn)
	jg.Add(upstreamWrite,conn)
	jg.Add(clientRead, conn)
	jg.Add(clientWrite, conn)

	jg.Start()

	time.Sleep(60 * time.Second)
	jg.Shutdown()
}
