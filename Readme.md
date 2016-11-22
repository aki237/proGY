# proGY
A miniature squid like intermediate local proxy server.

## Installation
Usual go get is suffice to get this server up and running.
```Go
go get -u github.com/aki237/proGY
```
If `$GOPATH/bin` is in your path then you are good to go. Else add it in your $PATH variable.

Now a binary release is made possible for systemd based linux distros (x86_64).
```shell
$ curl -L "https://git.io/vXanp" | sh
```
This will setup proGY for you automatically. Don't run this script as root. Answer the followup questions and you'll be good to go.


## Usage

```Shell
proGY
```

This starts a proxy server at the http://[local_listen_address]:[listen_port] and authenticating requests to the
remote proxy server with the given proxy credentials.
If the network traffic is to be monitored : add a -v to the command. THis make live logs to appear in the stdout.
[-v is not recomended : craps up the terminal]

If a LAN IP is assigned to the host PC ,say 192.168.1.100, then proGY can be run in that address, thus accessible
to all the devices in the network.
Actually I got sick of the commandline arguments. So write a config file ".progy" in json format that conatains the
following ...

```Json
{
    "listenaddress":":9999",
    "remoteproxyaddress":"<listenaddress>:<port>",
    "Creds":[
		{
			"username":"<username1>",
			"password":"<password1>"
		},
		{
			"username":"<username2>",
			"password":"<password2>"
		},
		{
			"username":"<username3>",
			"password":"<password3>"
		}
	],
	"domaincachefile" : "<cacheFileLocation>",
	"verbose":true,
	"loggerport" :  <SomePortAsInteger>
}
```

### Example 
```Json
{
    "listenaddress":":9999",
    "remoteproxyaddress":"134.8.9.13:80",
    "Creds":[
		{
			"username":"alanthicke",
			"password":"ohcanada"
		},
		{
			"username":"brobibs",
			"password":"milliondollaridea"
		},
		{
			"username":"awesomium",
			"password":"elementbybarney"
		}
    ],
	"domaincachefile" : "/Users/stinson/.cache/dnscache.pgy",
    "verbose":true,
    "loggerport" :  3030
}
```

## Domain Name Cache File
Domain Name Cachefile points to a [bolt](https://github.com/boltdb/bolt) database file. 
If you don't have the file already, bolt will create it automatically. This file is used to 
store the maps of domain names to their IP addresses. This will make the tunnelling a lot faster than
domain name lookup for every connection.

## Logger Service TCP Port
This should be mentioned so that, the logger service will run at that port(TCP).
**What is a Logger Service**
It is a TCP based Cross Process Communication Service, with Json as the transport format.
**Why is it used?**
I used proGY as a systemd service. Till v1.02.1, it was spamming journal logs with lots of data 
(Sent Bytes, Recieved Bytes). I used it only when internet seems to slouch a little bit. So I was only
expecting for errors when I open the journal logs.
**How to use it?**
When you connect to the TCP server running at specified port using some program (say telnet), it will spit out 
the json object of the connection made or closed. This can be used by other programs. Like for example *monitoring*

## !Important!
This branch is just testing and is inteded to work in any posix system which has the proc file system mounted.
The running system has to have the `ss` command(not netstat). I have been writing another package to remove this
external binary dependancy, replicating the functionalities of `ss` and `ps` commands in golang. Any help to
write those packages will be apperitiated.
