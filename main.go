package main

func main() {
	server := NewServer("192.168.3.104", 8888)
	server.Start()
}
