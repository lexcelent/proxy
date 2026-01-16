package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
)

var filterList []string

func init() {
	file, err := os.Open("list.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		filterList = append(filterList, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("error starting tcp server:", err)
	}
	defer listener.Close()
	fmt.Printf("listening on %s:%s\n", "localhost", "8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("error accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(localConn net.Conn) {
	defer localConn.Close()

	for {
		buf := make([]byte, 1500)
		n, err := localConn.Read(buf)
		if n > 0 {
			fmt.Printf("read %d bytes\n", n)
			// fmt.Printf("data: %s\n", buf)
		}
		if err != nil {
			fmt.Printf("error reading stream: %s\n", err)
			break
		}

		// Looking for HTTP-packet
		if res := bytes.Compare(buf[:7], []byte("CONNECT")); res != 0 {
			localConn.Close()
			return
		}

		localConn.Write([]byte("HTTP/1.1 200 OK\n\n"))

		// Process HTTP-packet
		address, host, port := getAddressFromStream(buf)

		// open new connection
		remoteConn, err := net.Dial("tcp", address)
		if err != nil {
			fmt.Println("error connecting:", err)
			return
		}
		defer remoteConn.Close()

		// filter logic
		// Here is processing ClientHello packet.
		if containsInList(string(host)) && string(port) == "443" {
			data := make([]byte, 1500)

			n, err := localConn.Read(data)
			if err != nil {
				fmt.Println("error reading data:", err)
			}

			// if isClientHello(data) {
			// fmt.Println("ClientHello detected")
			// }
			if n > 5 {
				data = fragmentate(data[5:n])
				remoteConn.Write(data)
			}
		}

		// Data exchange
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			io.Copy(remoteConn, localConn)
			wg.Done()
		}()

		go func() {
			io.Copy(localConn, remoteConn)
			wg.Done()
		}()

		wg.Wait()

		fmt.Println("end waiting")
	}
}

// fragmentate splits a packet into a random number of pieces.
// It demonstrates the server's ability to reassemble
// packets piece by piece.
func fragmentate(chunk []byte) []byte {
	var parts [][]byte

	for len(chunk) != 0 {
		// first byte array always 0x16 0x3 X X X 0x1
		chunkLen := rand.Intn(len(chunk)) + 1 // Random number of parts

		var buf bytes.Buffer
		buf.Write([]byte{22, 3}) // 0x16 0x0303 or 0x16 0x03 + randint

		// TLS version
		tls := rand.Intn(255)
		buf.WriteByte(byte(tls))

		// Packet length in bytes as big-endian
		endian := make([]byte, 2)
		binary.BigEndian.PutUint16(endian, uint16(chunkLen))
		buf.Write(endian)

		buf.Write(chunk[0:chunkLen])

		parts = append(parts, buf.Bytes())

		chunk = chunk[chunkLen:]
	}

	return bytes.Join(parts, []byte(""))
}

// isClientHello checks is it ClientHello packet
func isClientHello(buffer []byte) bool {
	if len(buffer) < 5 {
		return false
	}

	if buffer[0] != 0x16 {
		return false
	}

	if len(buffer) >= 6 && buffer[5] == 0x01 {
		return true
	}

	return false
}

// setReadBuffer sets readBuffer to new value
func setReadBuffer(conn net.Conn) {
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		err := tcpConn.SetReadBuffer(1500)
		if err != nil {
			fmt.Printf("failed to set read buffer: %s\n", err)
		}
	}
}

// getAddressFromStream takes host and port out of HTTP-packet
func getAddressFromStream(buffer []byte) (string, []byte, []byte) {
	// CONNECT www.google.com:443 HTTP/1.1
	firstLine := bytes.Split(buffer, []byte("\r\n"))[0]

	// [CONNECT www.google.com:443 HTTP/1.1]
	dataSplitted := bytes.Split(firstLine, []byte(" "))

	// [www.google.com 443]
	hostAndPort := bytes.Split(dataSplitted[1], []byte(":"))

	// www.google.com:443
	address := string(hostAndPort[0]) + ":" + string(hostAndPort[1])

	return address, hostAndPort[0], hostAndPort[1]
}

// containsInList checks if item exists in list.
// O(n) check depends on length of filterList
func containsInList(data string) bool {
	for _, elem := range filterList {
		if strings.Contains(data, elem) {
			return true
		}
	}

	return false
}
