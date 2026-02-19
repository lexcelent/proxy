package clienthello

import (
	"encoding/binary"
	"fmt"
)

// Parse takes TLS packet and shows first SNI Extension
func Parse(packet []byte) {
	TLSType := packet[0]
	TLSVersion := packet[1:3] // 2 bytes
	TLSLength := packet[3:5]
	HandshakeType := packet[5] // client hello
	HandshakeLength := packet[6:9]
	ClientTLSVersion := packet[9:11]

	// probably 32 bytes and not 28
	RandomBytes := packet[11:43]

	fmt.Println(TLSType, TLSVersion, TLSLength, HandshakeType, HandshakeLength, ClientTLSVersion, RandomBytes)

	SessionIDLength := packet[43]
	SessionID := packet[44 : 44+int(SessionIDLength)]

	fmt.Printf("SessionIDLength: %d\tSessionID: %v\n", SessionIDLength, SessionID)
	fmt.Printf("Real Length: %d\n", len(SessionID))

	// SESSION ID
	arr := packet[44+int(SessionIDLength):]

	// 2 bytes
	CipherSuitesLengthBytes := arr[0:2]
	CipherSuitesLength := binary.BigEndian.Uint16(CipherSuitesLengthBytes) // get cipher length

	CipherSuites := arr[2 : 2+CipherSuitesLength]

	fmt.Printf("CipherSuitesLengthBytes: %v\tCipherSuitesLength: %d\n", CipherSuitesLengthBytes, CipherSuitesLength)
	fmt.Printf("CipherSuites: %v\n", CipherSuites)

	arr = arr[2+CipherSuitesLength:]

	CompMethodsLength := arr[0]
	CompMethods := arr[1] // or add CompMethodsLength to get slice ?

	fmt.Printf("CompMethodsLength: %d\tCompMethods: %d\n", CompMethodsLength, CompMethods)

	// Extensions
	ExtensionsLength := binary.BigEndian.Uint16(arr[2:4]) // arr[2:4] to int

	arr = arr[4:100] // we don't need a lot of data

	fmt.Printf("ExtensionsLength: %d\n", ExtensionsLength)

	ExtensionType := arr[0:2] // type 00 00 == SNI

	ExtensionLength := binary.BigEndian.Uint16(arr[2:4])

	fmt.Printf("ExtensionType: %v\tExtensionLength: %d\n", ExtensionType, ExtensionLength)

	ServerNameListLength := arr[4:6]
	ServerNameType := arr[6] // 00 hostname
	ServerNameLengthBytes := arr[7:9]
	ServerNameLength := binary.BigEndian.Uint16(ServerNameLengthBytes)
	ServerName := arr[9 : 10+ServerNameLength]

	fmt.Printf("ServerNameListLength: %v\n", ServerNameListLength)
	fmt.Printf("ServerNameType: %d\n", ServerNameType)

	fmt.Printf("%s\n", ServerName)
}
