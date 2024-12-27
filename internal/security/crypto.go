package security

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// chatgpt
func compressECDSAPublicKey(pub *ecdsa.PublicKey) string {
	// Erstelle das Kompressionspräfix basierend auf der Y-Koordinate
	// 0x02, wenn Y gerade ist, 0x03, wenn Y ungerade ist
	prefix := byte(0x02)
	if pub.Y.Bit(0) == 1 {
		prefix = 0x03
	}

	// Füge das Präfix und die X-Koordinate zusammen
	compressed := append([]byte{prefix}, pub.X.Bytes()...)

	// Gib den Schlüssel als hexadezimale Zeichenkette zurück
	return hex.EncodeToString(compressed)
}

func GenerateShortPublicKey() string {
	// Verwende eine elliptische Kurve mit kleiner Schlüssellänge (P224)
	privateKey, err := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	if err != nil {
		fmt.Errorf("Fehler beim Generieren des Schlüssels: %v", err)
		panic(err)
	}

	publicKey := privateKey.PublicKey

	return compressECDSAPublicKey(&publicKey)
}
