package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"net"
	"encoding/json"
	random "math/rand"
	b64 "encoding/base64"
)

type Cred struct{
	Username string
	Password string
}

type Creds []Cred

type Config struct{
	Listenaddress string
	Remoteproxyaddress string
	ProxyCreds Creds
	Verbose bool
}

//A proxy represents a pair of connections and their state
type proxy struct {
	sentBytes     uint64
	receivedBytes uint64
	laddr, raddr  *net.TCPAddr
	lconn, rconn  *net.TCPConn
	erred         bool
	errsig        chan bool
	prefix        string
	encauth       []string
}

//Init variables
var matchid = uint64(0)
var connid = uint64(0)
var localAddr = flag.String("l", ":9999", "local address")
var remoteAddr = flag.String("r", "10.1.1.18:80", "Remote Proxy Address")
var authpair = flag.String("a","<username>:<password>","Proxy authentication details -- dont use quotes for it")
var verbose = flag.Bool("v", false, "display server actions")
var veryverbose = flag.Bool("vv", false, "display server actions and all tcp data")
var nagles = flag.Bool("n", false, "disable nagles algorithm")

//Main function to start the server
func main() {
	home := os.Getenv("HOME")
	flag.Parse()
	content,err := ioutil.ReadFile(home+"/.progy")
	if (err != nil){
		fmt.Println("Unable to open config file : Using defaults",err)
	}
	var conf Config
	err = json.Unmarshal(content,&conf)
	*localAddr = conf.Listenaddress
	*remoteAddr = conf.Remoteproxyaddress
	*verbose = conf.Verbose
	fmt.Printf("Proxying from %v to %v\n", *localAddr, *remoteAddr)	
	laddr, err := net.ResolveTCPAddr("tcp", *localAddr)
	check(err)
	raddr, err := net.ResolveTCPAddr("tcp", *remoteAddr)
	check(err)
	listener, err := net.ListenTCP("tcp", laddr)
	check(err)
	encauth := make([]string,0)
	for _,val := range conf.Creds {
		encauth = append(encauth,b64.StdEncoding.EncodeToString([]byte(val.Username+":"+val.Password)))
	}
	fmt.Println(encauth)

	if *veryverbose {
		*verbose = true
	}

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			fmt.Printf("Failed to accept connection '%s'\n", err)
			continue
		}
		connid++

		p := &proxy{
			lconn:    conn,
			laddr:    laddr,
			raddr:    raddr,
			erred:    false,
			errsig:   make(chan bool),
			prefix:   fmt.Sprintf("Connection #%03d ", connid),
			encauth:  encauth,
		}
		go p.start()
	}
}

//Logging function
func (p *proxy) log(s string, args ...interface{}) {
	if *verbose {
		log(p.prefix+s, args...)
	}
}

//Proxy error fuction
func (p *proxy) err(s string, err error) {
	if p.erred {
		return
	}
	if err != io.EOF {
		log(p.prefix+s, err)
	}
	p.errsig <- true
	p.erred = true
}


//Proxy Dial function
func (p *proxy) start() {
	defer p.lconn.Close()
	//connect to remote
	rconn, err := net.DialTCP("tcp", nil, p.raddr)
	if err != nil {
		p.err("Remote connection failed: %s\n", err)
		return
	}
	p.rconn = rconn
	defer p.rconn.Close()
	//nagles?
	if *nagles {
		p.lconn.SetNoDelay(true)
		p.rconn.SetNoDelay(true)
	}
	//display both ends
	p.log("Opened %s → %s\n", p.lconn.RemoteAddr().String(), p.rconn.RemoteAddr().String())
	//bidirectional copy
	go p.pipe(p.lconn, p.rconn)
	go p.pipe(p.rconn, p.lconn)
	//wait for close...
	<-p.errsig
	p.log("Closed (%d bytes sent, %d bytes recieved)\n", p.sentBytes, p.receivedBytes)
}


//Piping proxy requests to the remote
func (p *proxy) pipe(src, dst *net.TCPConn) {
	var f string
	islocal := src == p.lconn
	buff := make([]byte, 0xffff)
	for {
		n, err := src.Read(buff)
		if err != nil {
			p.err("Read failed '%s'\n", err)
			return
		}
		b := buff[:n]
		if islocal{
			var netstr string = string(b)
			if (strings.Contains(netstr,"User-Agent")){
				netstr = strings.Replace(netstr,"\nUser-Agent:","\nProxy-Authorization: Basic "+p.encauth[random.Intn(len(p.encauth))]+"\nUser-Agent:",1)
			}else{
				if(strings.Contains(netstr,"CONNECT")||strings.Contains(netstr,"GET")){
					netstr = strings.Replace(netstr,"\n","\nProxy-Authorization: Basic "+p.encauth[random.Intn(len(p.encauth))]+"\n",1)
				}
			}
			b = []byte(netstr)
			f = "Sent -> "
		}else{
			f = "Recv -> "
		}
		//show output
		if *veryverbose {
			if islocal{
				fmt.Println(string(b))
			}
		} else {
			if islocal{
				fmt.Println(f,n)
			}
		}
		//write out result
		n, err = dst.Write(b)
		if err != nil {
			p.err("Write failed '%s'\n", err)
			return
		}
		if islocal {
			p.sentBytes += uint64(n)
		} else {
			p.receivedBytes += uint64(n)
		}
	}
}

//helper functions

func check(err error) {
	if err != nil {
		log(err.Error())
		os.Exit(1)
	}
}

func log(f string, args ...interface{}) {
	fmt.Printf(f, args...)
}

//
func writeToFile(content string)  {
	f, err := os.OpenFile(os.Getenv("HOME")+"/progylog", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	content = content + "#INDICATOR#m\n------------------------------------\n\n===========================================================\n\n-----------------------------\n"
	_, err = f.WriteString(content)
	if  err != nil {
		panic(err)
	}
}
