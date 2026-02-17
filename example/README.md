# Example of usage

Step by step guide

1. Create `main.go`
2. Create server
3. Call server.ListenAndServe

The server is running and ready for TCP connections!

```
func main() {
	s := proxy.Server{}
	if err := s.ListenAndServe("tcp", ":8080"); err != nil {
		log.Fatalf("error: %s\n", err)
	}
}
```
