package main

import "net"

func ServerMain() {
	ln, err := net.Listen("tcp", fServerAddr)
	if err != nil {
		panic(err)
	}
	defer ln.Close()
}
