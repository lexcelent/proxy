package clienthello

import (
	"encoding/binary"
	"fmt"

	"golang.org/x/crypto/cryptobyte"
)

type TLSClientHello struct {
	TLSType    byte
	TLSVersion []byte
	TLSLength  []byte

	HandshakeType   byte
	HandshakeLength []byte

	TLSClientVersion []byte
	RandomBytes      []byte

	SessionIDLength byte
	SessionID       []byte

	CipherSuitesLength []byte
	CipherSuites       []byte

	CompMethodsLength byte
	CompMethods       byte

	ExtensionsLength []byte

	ExtType   []byte
	ExtLength []byte

	ServerNameListLength []byte
	ServerNameType       byte
	ServerNameLength     []byte
	ServerName           []byte
}

// Parse takes TLS packet and shows first SNI Extension
func Parse(packet []byte) error {
	p := cryptobyte.String(packet)

	if p.Empty() {
		return fmt.Errorf("error empty packet")
	}

	clientHello := &TLSClientHello{}

	/* Header */

	if !p.ReadUint8(&clientHello.TLSType) {
		return fmt.Errorf("error TLS Type")
	}

	if !p.ReadBytes(&clientHello.TLSVersion, 2) {
		return fmt.Errorf("error TLS Version")
	}

	if !p.ReadBytes(&clientHello.TLSLength, 2) {
		return fmt.Errorf("error TLS Length")
	}

	/* Handshake */

	if !p.ReadUint8(&clientHello.HandshakeType) {
		return fmt.Errorf("error Handshake type")
	}

	if !p.ReadBytes(&clientHello.HandshakeLength, 3) {
		return fmt.Errorf("error Handshake Length")
	}

	if !p.ReadBytes(&clientHello.TLSClientVersion, 2) {
		return fmt.Errorf("error TLS Client Version")
	}

	if !p.ReadBytes(&clientHello.RandomBytes, 32) {
		return fmt.Errorf("error TLS Random bytes")
	}

	/* Session info */

	if !p.ReadUint8(&clientHello.SessionIDLength) {
		return fmt.Errorf("error SessionID Length")
	}

	if !p.ReadBytes(&clientHello.SessionID, int(clientHello.SessionIDLength)) {
		return fmt.Errorf("error SessionID")
	}

	/* Cipher Suites */

	if !p.ReadBytes(&clientHello.CipherSuitesLength, 2) {
		return fmt.Errorf("error Cipher Suites Length")
	}

	if !p.ReadBytes(&clientHello.CipherSuites, int(binary.BigEndian.Uint16(clientHello.CipherSuitesLength))) {
		return fmt.Errorf("error Cipher Suites")
	}

	/* Compression */

	if !p.ReadUint8(&clientHello.CompMethodsLength) {
		return fmt.Errorf("error Compression Methods Length")
	}

	if !p.ReadUint8(&clientHello.CompMethods) {
		return fmt.Errorf("error Compression Methods")
	}

	/* Extensions */

	if !p.ReadBytes(&clientHello.ExtensionsLength, 2) {
		return fmt.Errorf("error Extensions Length")
	}

	/* Single Extension */

	// TODO: Wrap in loop to parse every extension

	if !p.ReadBytes(&clientHello.ExtType, 2) {
		return fmt.Errorf("error Extension Type")
	}

	if !p.ReadBytes(&clientHello.ExtLength, 2) {
		return fmt.Errorf("error Extension Length")
	}

	// Parse SNI

	// TODO: Wrap for every SNI

	if !p.ReadBytes(&clientHello.ServerNameListLength, 2) {
		return fmt.Errorf("error Server Name List Length")
	}

	if !p.ReadUint8(&clientHello.ServerNameType) {
		return fmt.Errorf("error Server Name Type")
	}

	if !p.ReadBytes(&clientHello.ServerNameLength, 2) {
		return fmt.Errorf("error Server Name Length")
	}

	if !p.ReadBytes(&clientHello.ServerName, int(binary.BigEndian.Uint16(clientHello.ServerNameLength))) {
		return fmt.Errorf("error Server Name")
	}

	fmt.Printf("Server Name: %s\n", clientHello.ServerName)

	return nil
}
