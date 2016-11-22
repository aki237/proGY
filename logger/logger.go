package logger

import (
	"encoding/json"
	"fmt"
	"net"
)

type tip struct {
	net.Conn
	Closed bool
}

var server net.Listener

var t tip

type connection struct {
	Process string `json:"process"`
	Thru    string `json:"proxyserver"`
	Host    string `json:"host"`
	ID      int    `json:"connid"`
	Opening bool   `json:"opening"`
}

func Init(port int) error {
	var err error
	server, err = net.Listen("tcp", ":"+fmt.Sprint(port))
	if err != nil {
		return err
	}
	go makeSingle()
	return err
}

func makeSingle() {
	t = tip{Closed: true}
	conn, err := server.Accept()
	if err != nil {
		return
	}
	t.Closed = false
	t.Conn = conn
}

func Log(process, proxyServer, host string, connid int, opening bool) {
	if t.Closed {
		return
	}
	conn := &connection{
		Process: process,
		Thru:    proxyServer,
		Host:    host,
		ID:      connid,
		Opening: opening,
	}
	b, err := json.Marshal(conn)
	if err != nil {
		return
	}
	_, err = t.Write([]byte(string(b) + "\n"))
	if err != nil {
		t.Close()
		t.Closed = true
		go makeSingle()
	}
}
