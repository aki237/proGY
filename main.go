package main

import (
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	random "math/rand"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/aki237/proGY/logger"
)

// Cred is a struct for holding the proxy authentication credentials (username, password)
type Cred struct {
	Username           string // Username : variable for holding username
	Password           string // Password : variable for holding password
	Remoteproxyaddress string // Remoteproxyaddress : remote proxy address
	Encauth            string
}

// Creds is of the type slice of Cred
type Creds []Cred

// Config is the struct used to hold the configuration read from the configuration file
type Config struct {
	Listenaddress   string // ListenAddress : local listen address
	Creds           Creds  // Creds : Slice of all credentials for the specified remote proxy address
	Verbose         bool   // Verbose : Whether to be verbose about the output
	Domaincachefile string // Size of the DNS Cache to be saved during runtime
	Loggerport      int    // What port to run the Logger at
	ControlSocket   string // At which file to run the unix socket for controlling proGY
	*sync.Mutex
}

func (c *Config) Reloader(fileChannel chan string) {
	for {
		filename := <-fileChannel
		c.Lock()
		temp := parseConfig(filename)
		c.Creds = temp.Creds
		c.Verbose = temp.Verbose
		c.Domaincachefile = temp.Domaincachefile
		c.Unlock()
	}
}

//A proxy represents a pair of connections and their state
type proxy struct {
	sentBytes     uint64
	receivedBytes uint64
	laddr, raddr  *net.TCPAddr
	lconn, rconn  *net.TCPConn
	erred         bool
	errsig        chan bool
	encauth       string
	site          string
	process       string
	connid        int
}

var connid = 0
var verbose bool
var dnscache Cache

func parseConfig(filename string) Config {
	home := os.Getenv("HOME")
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log("Unable to open config file : Using defaults %s\n", err)
		os.Exit(2)
	}
	var conf Config
	err = json.Unmarshal(content, &conf)
	if err != nil {
		log("Error : %s\n", err)
		os.Exit(2)
	}
	if conf.Domaincachefile == "" {
		conf.Domaincachefile = home + "/.cache/dnscache.pgy"
	}
	for i, val := range conf.Creds {
		conf.Creds[i].Encauth = b64.StdEncoding.EncodeToString([]byte(val.Username + ":" + val.Password))
	}
	return conf
}

func (conf *Config) getProxyStruct(conn *net.TCPConn) *proxy {
	localAddr := conf.Listenaddress
	currentProxy := conf.Creds[random.Intn(len(conf.Creds))]
	remoteAddr := currentProxy.Remoteproxyaddress
	verbose = conf.Verbose
	laddr, err := net.ResolveTCPAddr("tcp", localAddr)
	check(err)
	raddr, err := net.ResolveTCPAddr("tcp", remoteAddr)
	check(err)
	proxyStruct := &proxy{
		lconn:   conn,
		laddr:   laddr,
		raddr:   raddr,
		erred:   false,
		errsig:  make(chan bool),
		encauth: currentProxy.Encauth,
		connid:  connid,
	}
	return proxyStruct
}

//Main function to start the server
func main() {
	home := os.Getenv("HOME")
	wordPtr := flag.String("config", home+"/.progy", "Configuration file to be provided for proGY")
	flag.Parse()
	conf := parseConfig(*wordPtr)
	conf.Mutex = &sync.Mutex{}
	laddr, err := net.ResolveTCPAddr("tcp", conf.Listenaddress)
	check(err)
	listener, err := net.ListenTCP("tcp", laddr)
	check(err)
	dnscache, err = NewCache(conf.Domaincachefile)
	check(err)
	err = logger.Init(conf.Loggerport)
	check(err)
	fileChannel := make(chan string)
	go listenUnixControl(conf.ControlSocket, fileChannel)
	go conf.Reloader(fileChannel)
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			fmt.Printf("Failed to accept connection '%s'\n", err)
			continue
		}
		connid++
		p := conf.getProxyStruct(conn)
		go p.start()
	}
}

//Logging function
func (p *proxy) log(s string, args ...interface{}) {
	log(s, args...)
}

//Proxy error fuction
func (p *proxy) err(s string, err error) {
	if p.erred {
		return
	}
	if err != io.EOF {
		log(s, err.Error())
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
		log("Remote connection failed: %s\n", err)
		return
	}
	p.rconn = rconn
	defer p.rconn.Close()
	//nagles?
	p.lconn.SetNoDelay(true)
	p.rconn.SetNoDelay(true)
	//display both ends
	host := strings.Split(p.lconn.RemoteAddr().String(), ":")[0]
	port := strings.Split(p.lconn.RemoteAddr().String(), ":")[1]
	o, _ := exec.Command("/sbin/ss", "-ntp", "src", host, "sport", "=", ":"+port).Output()
	process := "systemAt:" + host
	//fmt.Println(string(o))
	splitted := strings.Split(string(o), "users:((\"")
	if len(splitted) > 1 {
		process = strings.Split(splitted[1], "\",pid=")[0]
	}
	p.process = process
	//bidirectional copy
	go p.pipe(p.lconn, p.rconn)
	go p.pipe(p.rconn, p.lconn)
	//wait for close...
	<-p.errsig
	go logger.Log(p.process, p.raddr.IP.String(), p.site, connid, logger.STATUS_CLOSED, p.receivedBytes)
}

//Piping proxy requests to the remote
func (p *proxy) pipe(src, dst *net.TCPConn) {
	//var f string
	islocal := src == p.lconn
	buff := make([]byte, 0xffff)
	var n int
	for {
		var err error
		n, err = src.Read(buff)
		if n == 0 {
			p.err("%s", errors.New(""))
			return
		}
		if err != nil {
			p.err("Read failed '%s'\n", err)
			return
		}
		b := buff[:n]
		netstr := string(b)
		var host string
		if islocal && (strings.Contains(netstr, "CONNECT") ||
			strings.Contains(netstr, "GET") ||
			strings.Contains(netstr, "HEAD") ||
			strings.Contains(netstr, "POST")) {
			netstr = strings.Replace(netstr, "\n", "\nProxy-Authorization: Basic "+p.encauth+"\n", 1)
			reqtype := strings.Split(netstr, "\n")[0]
			splitted := strings.Split(reqtype, " ")
			if len(splitted) > 1 {
				host = splitted[1]
				go logger.Log(p.process, p.raddr.IP.String(), host, connid, logger.STATUS_OPENED, 0)
				if strings.Contains(splitted[0], "CONNECT") {
					if strings.Contains(host, ":") {
						host = strings.Split(host, ":")[0]
					}
					ip, err := dnscache.LookupIP(host)
					if err != nil {
						fmt.Println(err)
						n, err = dst.Write([]byte(netstr))
						if err != nil {
							p.err("Error : ", err)
							return
						}
						return
					}
					IP := ip
					netstr = strings.Replace(netstr, host, IP, 1)
				}
				p.site = host
			}
		}
		n, err = dst.Write([]byte(netstr))
		if err != nil {
			p.err("Unable To connect : %s\n", err)
			return
		}
		if islocal {
			p.sentBytes += uint64(n)
		} else {
			p.receivedBytes += uint64(n)
			go logger.Log(p.process, p.raddr.IP.String(), p.site, connid, logger.STATUS_INPROCESS, p.receivedBytes)
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
