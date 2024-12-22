# Gonet
### Spun up simple ascii chat over wireguard p2p

## Usage:

```yaml
gonet <lobby/room/key>
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

### links
https://tailscale.com/blog/how-nat-traversal-works
https://coder.com/docs/admin/networking/stun
https://fly.io/blog/ssh-and-user-mode-ip-wireguard/
https://fly.io/blog/our-user-mode-wireguard-year/
https://github.com/Schachte/userspace-wireguard-tunnels
https://ryan-schachte.com/blog/userspace_wireguard_tunnels/ <---- good one!
https://stackoverflow.com/questions/58129995/binding-a-udp-port-how-long-does-the-binding-persist-in-a-nat-environment
https://www.digitalocean.com/community/tutorials/how-to-create-a-point-to-point-vpn-with-wireguard-on-ubuntu-16-04#creating-an-initial-configuration-file
