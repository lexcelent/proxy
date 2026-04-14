# Proxy

Simple Golang HTTP-proxy server


# Getting started

It's simple. You can find "step by step" guide in `./example` folder

Example of `main.go` file:

```
package main

import (
	"log"

	"github.com/lexcelent/proxy"
)

func main() {
	s := proxy.Server{}
	if err := s.ListenAndServe("tcp", ":8080"); err != nil {
		log.Fatalf("error: %s\n", err)
	}
}
```

## DEV branch

`dev` contains experimental features.  
Currently it contains only Statistics (total bytes read)

## TODO

- HTTPS
- logger
- statistics
