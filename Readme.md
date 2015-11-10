# proGY
A miniature squid like intermediate local proxy server.

## Installation
Usual go get is suffice to get this server up and running.
```Go
go get -u github.com/aki237/proGY
```
If `$GOPATH/bin` is in your path then you are good to go. Else add it in your $PATH variable.

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
    "listenaddress":"[local_listen_address]:[port]",
    "remoteproxyaddress":"[remote_proxy_address]:[remote_proxy_port]",
    "username":"[remote_proxy_username]",
    "password":"[remote_proxy_password]",
    "verbose":[true or false without quotes]
}
```

### Example 
```Json
{
    "listenaddress":"127.0.0.1:9999",
    "remoteproxyaddress":"10.8.8.90:80",
    "username":"alanthick",
    "password":"canadian",
    "verbose":true
}
```
