# Telnet reverse proxy

This server allow two clients to exchange text data.
One is deemed the client, the other one the server, but both connect on the same TCP port.
The two hosts use a 'channel' identifier. Once a pair of client/server has connected, another pair can use the same channel name.
## Usage
The server has to connect first. It will wait until there is a client using the same channel name
### Server
```shell script
telnet proxy.ribes.ovh 23
> s_channel1
registred as server in channel _channel1
# client connects
hi
> howdoyoudo
```

### Client
```shell script
telnet proxy.ribes.ovh 23
#server connects
> c_channel1
registred as client in channel
> hello
howdoyoudo
```

### Closing connection
When one host disconnects, both are disconnected.
One can send 'EOF' to be gracefully disconnected. It will inform the other side