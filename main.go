package main

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/pion/stun"
	"golang.org/x/sys/unix"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/tun/netstack"
	"log"
	"net"
	"net/netip"
	"os"
	"sync"
	"syscall"
	"time"
)

type WireguardConfig struct {
	PrivateKey string
	PublicKey  string
	AllowedIPs string
	Endpoint   string
	Port       int
}

type Wireguard struct {
	config WireguardConfig
}

func (w *Wireguard) GenerateTUN(localAddresses []netip.Addr, dnsAddresses []netip.Addr, mtu *int) (tun.Device, *netstack.Net, error) {
	defaultMtu := 1500
	if mtu == nil {
		mtu = &defaultMtu
	}

	tun, tnet, err := netstack.CreateNetTUN(
		localAddresses,
		dnsAddresses,
		*mtu,
	)
	return tun, tnet, err
}

func (w *Wireguard) CreateDevice(tunDevice tun.Device, logLevel int) (*device.Device, error) {
	dev := device.NewDevice(
		tunDevice,
		conn.NewDefaultBind(),
		device.NewLogger(logLevel, ""),
	)
	if dev == nil {
		return nil, fmt.Errorf("Failed to create device")
	}
	return dev, nil
}

func startUdpServer(port int) {
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

	// Read from UDP listener in endless loop
	for {
		var buf [512]byte
		_, _, err := conn.ReadFromUDP(buf[0:])
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Print("> ", string(buf[0:]))

		// Write back the message over UPD
		// conn.WriteToUDP([]byte("Hello UDP Client\n"), addr)
	}
}

func startUdpClient() {
	time.Sleep(1 * time.Second)
	if len(os.Args) == 1 {
		fmt.Println("Please provide host:port to connect to")
		os.Exit(1)
	}

	// Resolve the string address to a UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", os.Args[2])

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Dial to the address with UDP
	conn, err := net.DialUDP("udp", nil, udpAddr)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Send a message to the server
	_, err = conn.Write([]byte("Hello UDP Server\n"))
	fmt.Println("send...")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getAddress() int {
	// Parse a STUN URI
	u, err := stun.ParseURI("stun:stun.schlund.de")
	if err != nil {
		panic(err)
	}

	// Creating a "connection" to STUN server.
	c, err := stun.DialURI(u, &stun.DialConfig{})
	if err != nil {
		panic(err)
	}
	// Building binding request with random transaction id.
	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)

	var port int

	// Sending request to STUN server, waiting for response message.
	if err := c.Do(message, func(res stun.Event) {
		if res.Error != nil {
			panic(res.Error)
		}
		// Decoding XOR-MAPPED-ADDRESS attribute from message.
		var xorAddr stun.XORMappedAddress
		if err := xorAddr.GetFrom(res.Message); err != nil {
			panic(err)
		}
		fmt.Println("your IP is", xorAddr.IP)
		port = xorAddr.Port
	}); err != nil {
		panic(err)
	}

	return port
}

func main() {
	//cmd.Execute()

	p := getAddress()

	waitGroup := sync.WaitGroup{}

	waitGroup.Add(1)
	go startUdpServer(p)
	waitGroup.Add(1)
	go startUdpClient()

	waitGroup.Wait()
	os.Exit(0)

	preferredMTU := 1500
	wg := Wireguard{}
	tundv, _, err := wg.GenerateTUN(
		[]netip.Addr{netip.MustParseAddr("10.0.0.3")},
		[]netip.Addr{netip.MustParseAddr("1.1.1.1")},
		&preferredMTU)
	if err != nil {
		log.Panic("Failed to create TUN device:", err)
	}

	dev, err := wg.CreateDevice(tundv, device.LogLevelVerbose)
	if err != nil {
		log.Panic("Failed to create WireGuard device")
	}

	wgConfig := `
private_key=%s
public_key=%s
allowed_ip=0.0.0.0/0
endpoint=<PUBLIC_IP_ADDRESS>:51820
`

	err = dev.IpcSet(fmt.Sprintf(wgConfig, base64ToHex("<PRIVATE_KEY>"), base64ToHex("<PUBLIC_KEY>")))

	if err != nil {
		log.Panic("Failed to set WireGuard configuration:", err)
	}

	// bring up the Wireguard device
	err = dev.Up()
	if err != nil {
		log.Panic("Failed to bring up WireGuard device:", err)
	}

	fmt.Println("Connected to WireGuard server")
}

func base64ToHex(base64Key string) string {
	decodedKey, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		log.Panic("Failed to decode base64 key:", err)
	}
	hexKey := hex.EncodeToString(decodedKey)
	return hexKey
}
