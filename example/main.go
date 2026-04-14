package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/lexcelent/proxy"
)

var (
	port = flag.String("p", "8080", "proxy port")
)

func main() {
	flag.Parse()

	s := proxy.New(true)
	fmt.Printf("Server has been started on port %s\n", *port)
	if err := s.ListenAndServe("tcp", fmt.Sprintf(":%s", *port)); err != nil {
		log.Fatalf("error: %s\n", err)
	}
}
