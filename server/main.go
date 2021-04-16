package main

import("fmt")

func main() {
	fmt.Println("hello")
	server := NewServer("127.0.0.1",8888)
	fmt.Println("hello")
	server.Start()

}