package main

import (
	"log"
	"math/rand"
	"net"
	"sync/atomic"
	"time"

	squares "github.com/iBug/Squares-go"
)

type ClientInfo struct {
	id   int
	conn net.Conn
}

type ClientMessage struct {
	num int
	m   any
}

const IDRANGE = 1 << 30

var (
	waitingRoom []ClientInfo
	wrCount     int

	gameOngoing atomic.Bool

	r1 = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func handleClient(num int, ci ClientInfo, ch chan<- ClientMessage) {}

func serverGame(ch <-chan ClientInfo) {
	for wrCount < squares.NPLAYERS {
		ci := <-ch
		waitingRoom[wrCount] = ci
		wrCount++
	}

	cm := make(chan ClientMessage)
	for i := range waitingRoom {
		go handleClient(i, waitingRoom[i], cm)
	}
	resetWaitingRoom()

	gameOngoing.Store(true)
	defer gameOngoing.Store(false)
}

func resetWaitingRoom() {
	waitingRoom = make([]ClientInfo, squares.NPLAYERS)
	wrCount = 0
}

func serverMain() {
	resetWaitingRoom()

	ln, err := net.Listen("tcp", fServerAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	ch := make(chan ClientInfo, 1)
	go serverGame(ch)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			conn.Close()
			continue
		}
		ch <- ClientInfo{
			id:   r1.Intn(IDRANGE) + 1,
			conn: conn,
		}
	}
}
