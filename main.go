package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
)

// flags
var (
	port = flag.String("p", "8080", "proxy port")
)

const HTTPSPort string = "443"

var (
	filterList []string
	blockList  []string
)

func filterListInit() {
	filename := "list.txt"

	_, err := os.Stat(filename)
	if errors.Is(err, os.ErrNotExist) {
		fmt.Printf("there is no filters\n")
		return
	} else if err != nil {
		log.Fatalf("error check file exists: %s", err)
	}

	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("try to open filter list: %s\n", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		// Комментарии в список не добавляем
		if strings.HasPrefix(scanner.Text(), "#") {
			continue
		}

		// Пустые строки игнорируем
		if len(scanner.Text()) == 0 {
			continue
		}

		filterList = append(filterList, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func blockListInit() {
	filename := "block.txt"

	_, err := os.Stat(filename)
	if errors.Is(err, os.ErrNotExist) {
		fmt.Printf("there is no block filters\n")
		return
	} else if err != nil {
		log.Fatalf("error check file exists: %s", err)
	}

	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("try to open filter list: %s\n", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		blockList = append(blockList, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	filterListInit()
	blockListInit()
}

func main() {
	flag.Parse()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", *port))
	if err != nil {
		log.Fatal("error starting tcp server:", err)
	}
	defer listener.Close()
	fmt.Printf("listening on %s:%s\n", "127.0.0.1", *port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("error accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}

// isHTTPPacket checks is it HTTP packet or not
func isHTTPPacket(b []byte) bool {
	if res := bytes.Compare(b[:7], []byte("CONNECT")); res != 0 {
		return false
	}

	return true
}

func handleConnection(localConn net.Conn) {
	defer localConn.Close()

	buf := make([]byte, 1500)
	n, err := localConn.Read(buf)
	if n > 0 {
		fmt.Printf("read %d bytes\n", n)
		// fmt.Printf("data: %s\n", buf)
	}
	if err != nil {
		fmt.Printf("error reading stream: %s\n", err)
		return
	}

	// Looking for HTTP-packet. If not - skip
	if !isHTTPPacket(buf) {
		return
	}

	// Process HTTP-packet
	address, host, port := getAddressFromStream(buf)

	// block sites you don't wanna see
	if containsInList(string(host), blockList) && string(port) == HTTPSPort {
		return
	}

	localConn.Write([]byte("HTTP/1.1 200 OK\n\n"))

	// open new connection
	remoteConn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("error connecting:", err)
		return
	}
	defer remoteConn.Close()

	// filter logic
	// Here is processing ClientHello packet.
	if containsInList(string(host), filterList) && string(port) == HTTPSPort {
		data := make([]byte, 1500)

		n, err := localConn.Read(data)
		if err != nil {
			fmt.Println("error reading data:", err)
		}

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

// getAddressFromStream takes host and port out of HTTP-packet
func getAddressFromStream(buffer []byte) (string, []byte, []byte) {
	// CONNECT www.google.com:443 HTTP/1.1\r\n
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
// O(n) check depends on length of list
func containsInList(data string, list []string) bool {
	// TODO: Is it possible make search O(1) using map ?
	for _, elem := range list {
		if strings.Contains(data, elem) {
			return true
		}
	}

	return false
}
