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
proGY -l="[local_listen_address]:[listen_port]" -r="[proxy_address]:[port]" -a=[Proxy_Username]:[Proxy_Password]
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
	    "password":"<password1>",
	},
	{
	    "username":"<username2>",
	    "password":"<password2>",
	},
	{
	    "username":"<username3>",
	    "password":"<password3>",
	}
    ],
    "verbose":true
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
	    "password":"ohcanada",
	},
	{
	    "username":"brobibs",
	    "password":"milliondollaridea",
	},
	{
	    "username":"awesomium",
	    "password":"elementbybarney",
	}
    ],
    "verbose":true
}
```
