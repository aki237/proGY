package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func listenUnixControl(socketName string, fileChannel chan string) {
	if socketName == "" {
		socketName = "/tmp/proGY-control"
	}
	if _, err := os.Stat(socketName); err == nil {
		if os.Remove(socketName) != nil {
			fmt.Println(err)
			os.Exit(2)
		}
	}
	ln, err := net.ListenUnix("unix", &net.UnixAddr{socketName, "unix"})
	if err != nil {
		log("%s\n", err)
		os.Exit(2)
	}
	for {
		conn, _ := ln.Accept()
		go func(conn net.Conn) {
			b := bufio.NewReader(conn)
			s, e := b.ReadString('\n')
			if e != nil {
				return
			}
			s = strings.TrimSpace(s)
			splitted := strings.Split(s, " ")
			if len(splitted) < 2 {
				conn.Close()
				return
			}
			if splitted[0] != "RELOAD" {
				conn.Close()
				return
			}
			filename := ""
			for i, val := range splitted[1:] {
				filename += val
				if i != len(splitted[1:])-1 {
					filename += " "
				}
			}
			fileChannel <- filename
			conn.Write([]byte("SUCCESS\n"))
			conn.Close()
			return
		}(conn)
	}
}
