package wg

import (
	"fmt"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
	"log"
	"net/http"
	"os"
	"os/exec"
)

var serverPrivateKey = "uDBw2KjL2YZlUrt4DuPxP7kWNQySMI3OLjGRPCghcmM="
var serverPublicKey = "39bUSCHAxRTdas7CKGwE9xDeKuPvQF+n9O8gEGPZdxg="

var clientPrivateKey = "UMntopd6v8lCpgmwHpTLliIqtXqDgmCgRiovh3k4A38="
var clientPublicKey = "JGq1vBxd/HhPP/z9373bw6WTyE5UcJCT6WpSxkz4by4="

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

// (Optional) PostUp/PostDown für Routing und NAT (z. B. Internetzugang für Clients)
// PostUp = iptables -A FORWARD -i %i -j ACCEPT; iptables -A FORWARD -o %i -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
// PostDown = iptables -D FORWARD -i %i -j ACCEPT; iptables -D FORWARD -o %i -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE

func Run() {
	tunDevice, err := tun.CreateTUN("np0", device.DefaultMTU)

	connection := conn.NewDefaultBind()

	wgDevice := device.NewDevice(
		tunDevice,
		connection,
		device.NewLogger(device.LogLevelError, ""),
	)

	if err != nil {
		log.Panic("Failed to create WireGuard device")
	}

	serverIp := "10.0.0.1/24"

	config := `
[Interface]
# Private Key des Servers
PrivateKey = %s
# IP-Adresse des Servers im privaten Netzwerk
Address = %s
# Port für eingehende WireGuard-Verbindungen
ListenPort = 51820

# Erlaube IP-Weiterleitung
SaveConfig = true

# Beispiel Peer
[Peer]
# Der Public Key des Clients muss hier hinzugefügt werden
PublicKey = %s
# Erlaubte IPs für diesen Peer
AllowedIPs = 10.0.0.3/32
`
	fmt.Println(fmt.Sprintf(config, serverIp, serverPrivateKey, clientPublicKey))

	err = wgDevice.IpcSet(fmt.Sprintf(config, serverIp, serverPrivateKey, clientPublicKey))

	if err != nil {
		log.Panic("Failed to set WireGuard configuration:", err)
	}

	// bring up the Wireguard device
	err = wgDevice.Up()
	if err != nil {
		log.Panic("Failed to bring up WireGuard device:", err)
	}

	cmd := exec.Command("ip", "addr", "add", serverIp, "dev", "np0")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()

	if err != nil {
		panic(err)
	}

	go spawnHttpTestServer("10.0.0.1:1989")

	select {}
}

func Join() {
	tunDevice, err := tun.CreateTUN("np1", device.DefaultMTU)
	if err != nil {
		log.Panic("Failed to create TUN device:", err)
	}

	connection := conn.NewDefaultBind()

	wgDevice := device.NewDevice(
		tunDevice,
		connection,
		device.NewLogger(device.LogLevelError, ""),
	)

	defer wgDevice.Close()

	config := `
[Interface]
# Private Key des Servers
PrivateKey = %s
# IP-Adresse des Servers im privaten Netzwerk
Address = 10.0.0.3/32

[Peer]
PublicKey = %s
Endpoint = 192.168.188.100:51820
AllowedIPs = 10.0.0.1/32
PersistentKeepalive = 21
`

	fmt.Println(fmt.Sprintf(config, clientPrivateKey, serverPublicKey))

	err = wgDevice.IpcSet(config)
	if err != nil {
		log.Panic("Failed to configure WireGuard device:", err)
	}

	// Gerät hochfahren
	err = wgDevice.Up()
	if err != nil {
		log.Panic("Failed to bring up WireGuard device:", err)
	}

	// IP-Adresse an die TUN-Schnittstelle binden
	cmd := exec.Command("ip", "addr", "add", "10.0.0.3/32", "dev", "np1")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		log.Panic("Failed to set IP address on interface:", err)
	}

	fmt.Println("Erfolgreich mit dem WireGuard-Netzwerk verbunden!")

	select {}
}

func spawnHttpTestServer(ipAndPort string) {
	// Einfache Handler-Funktion
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hallo, dies ist ein einfacher HTTP-Server in Go!")
	})

	err := http.ListenAndServe(ipAndPort, nil)
	if err != nil {
		fmt.Printf("Fehler beim Starten des Servers: %v\n", err)
	}
}
