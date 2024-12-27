# Gonet
### Spun up simple ascii chat over wireguard p2p

## Usage:

```yaml
gonet <lobby/room/key>
```

## Signalling or finding lobbies

The host shares it's publickey to its clients. The publickey is alsow the mqtt topic where the host is listening on.
The host is the only one, that can publish signed messages on that channel. So its authenticity is verifiable.

The client publishes join information (name, publickey, ...) on the channel. The metadata is signed with its own publickey.

The host can verifiy the publickey is signed by the user and adds the publickey to its configuration. (allow access)
The host then publishes where the client can join as a signed message.

The client receives the information, verifies its signature and initates a connection with the provided information

EDIT: The publickey does NOT need to be the name of the channel/room/id. The host can craft a message, that contains the publickey and sign the message itself with
this publickey. The channel/room/id needs to be a substring of a secure hash of the host's publickey. 

A client can generate the hash based on the provided publickey and verify the signature.

`gonet join cf83e1`

SHA-512 Hash of a public key
cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e

example client request and host response:

```yaml
version: 1
---
room: cf83e1 <---- used for sanity checks
publicKey: "JGq1vBxd/HhPP/z9373bw6WTyE5UcJCT6WpSxkz4by4="
action: "join" <---- desired action
username: johndoe
---
signature: cff5ccf320b9bec2d7605d3b6d844c01
```

```yaml
version: 1
---
room: cf83e1 <---- used for sanity checks
peer: "10.0.0.23" <---- assigned ip address
publickey: "39bUSCHAxRTdas7CKGwE9xDeKuPvQF+n9O8gEGPZdxg="
method: "SHA-512" <---- verification method
address: "109.42.114.166:56112" <---- public reachable nat hole, to initiate wireguard connection
---
signature: 58efdfcc74a98a8b4d63a7f48ed4bf27
```

todo: This diagram does not show the public channel on which messages are distributed
```
----joins room---->     [Client]                                                           [Host]
                           ||       -----------------sendSignedMessage------------------>    || 
                           ||                                                                || (verifies signature)                                                              
                           ||                                                                || (adds client to config) 
                           ||       <----------------respondsWithSignedMessage-----------    ||
     (verifies signature)  ||
     (initiates join)      ||
```

## Add a password to a connection. (idea)

To only allow certain clients, you could share a password via a secure channel and ask for a hash of it, after a network is identified and before a client joins:

```
// The traffic is now on a direct connection between client and host
----initiates wg join---->     [Client]                                                           [Host]
                           ||       -----------------sendHashOfSharedPassword------------------>    || 
                           ||                                                                       || (verifies pw)                                                              
                           ||                                                                       || (allows further traffic) 
                           ||       <----------------respondsWithAck-----------                     ||
     (joins network)       ||
```

## Debugging

### SO_REUSEADDR / SO_REUSEPORT

```go
lc := net.ListenConfig{
  Control: func(network, address string, c syscall.RawConn) error {
    fmt.Println(network, address)
    var opErr error
    err := c.Control(func(fd uintptr) {
      opErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
      opErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
    })
    if err != nil {
      return err
    }
    return opErr
  },
}

lp, err := lc.ListenPacket(context.Background(), "udp", fmt.Sprintf("0.0.0.0:%d", port))

if err != nil {
panic(err)
}

conn := lp.(*net.UDPConn)
```

### stunclient
https://gist.github.com/zziuni/3741933
https://www.stunprotocol.org/
https://unix.stackexchange.com/a/698441
``
brew install stuntman
``

### nat behavior discovery
https://github.com/pion/stun/tree/master/cmd/stun-nat-behaviour

My nat (voadfone lte) has the following properties:
> address and port dependent This is the strictest of the three. Your NAT will only allow return traffic from exactly where you sent your UDP packet. Using this is not recommended, even if you configure mapping behavior correctly, because it will work poorly when the other NAT is misconfigured (fairly common).


### turn server
https://github.com/pion/turn


### ice test
https://icetest.info/

### using open dht instead of signaling server
https://github.com/manuels/wireguard-p2p
https://stackoverflow.com/questions/60425311/how-to-connect-to-peers-in-opendht

### using dns txt cache as distributed storage for metdata instead of signaling server
https://github.com/benjojo/dnsfs/blob/master/dnsfs/dns.go
https://blog.benjojo.co.uk/post/dns-filesystem-true-cloud-storage-dnsfs

Query some data:
``dig +short _poem.rednafi.com TXT | sed 's/[\" ]//g' | base64 -d``

Will return

```
I was angry with my friend;
I told my wrath, my wrath did end.
I was angry with my foe:
I told it not, my wrath did grow.

And I watered it in fears,
Night & morning with my tears:
And I sunned it with smiles,
And with soft deceitful wiles.

And it grew both day and night.
Till it bore an apple bright.
And my foe beheld it shine,
And he knew that it was mine.

And into my garden stole,
When the night had veiled the pole;
In the morning glad I see;
My foe outstretched beneath the tree.
```

### using mqtt as signaling server
https://www.emqx.com/en/blog/how-to-use-mqtt-in-golang
https://www.hivemq.com/mqtt/public-mqtt-broker/

### using etcd as shared brain
https://pkg.go.dev/go.etcd.io/etcd/server/v3/embed
https://github.com/etcd-io/etcd/tree/main/client/v3 <--- use wg subnet to define all possible clients

### pastebin as handshake / signaling server
https://pastebin.com/


./gonet join <roomId>
./gonet create <roomid>

### links
https://tailscale.com/blog/how-nat-traversal-works
https://coder.com/docs/admin/networking/stun
https://fly.io/blog/ssh-and-user-mode-ip-wireguard/
https://fly.io/blog/our-user-mode-wireguard-year/
https://github.com/Schachte/userspace-wireguard-tunnels
https://ryan-schachte.com/blog/userspace_wireguard_tunnels/ <---- good one!
https://stackoverflow.com/questions/58129995/binding-a-udp-port-how-long-does-the-binding-persist-in-a-nat-environment
https://www.digitalocean.com/community/tutorials/how-to-create-a-point-to-point-vpn-with-wireguard-on-ubuntu-16-04#creating-an-initial-configuration-file
https://security.stackexchange.com/a/119843 <---- awesome asymmetric encryption information
https://github.com/libp2p/go-libp2p
https://github.com/webmeshproj/webmesh
https://github.com/pojntfx/weron <------ awesome p2p webrtc wireguard alternative
