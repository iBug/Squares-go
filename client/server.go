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

func handleClient(num int, ci ClientInfo, ch chan<- ClientMessage) {
	for {
		msg, err := RecvMsg(ci.conn)
		if err != nil {
			log.Println(err)
			break
		}
		ch <- ClientMessage{num, msg}
	}
}

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
	game.Reset()
	// GameLoop:
	for {
		select {
		case cm := <-cm:
			num := cm.num
			var res any
			switch req := cm.m.(type) {
			case ConnectReq:
				res = ConnectRes{
					Id:       waitingRoom[num].id,
					PlayerId: num,
					Game:     *game,
				}
			case MoveReq:
				if num == game.ActivePlayer {
					pos := squares.Coord{req.Pos[0], req.Pos[1]}
					if !game.TryInsert(req.ShapeId, req.Rotation, pos, num, game.FirstRound) {
						res = MoveRes{
							Ok:           false,
							ActivePlayer: game.ActivePlayer,
						}
						break
					}
					game.Insert(req.ShapeId, req.Rotation, pos, num)
					game.AfterMove()
					res = MoveRes{
						Ok:           true,
						ActivePlayer: game.ActivePlayer,
					}

					for i := 0; i < squares.NPLAYERS; i++ {
						if i == num {
							continue
						}
						SendMsg(waitingRoom[i].conn, OtherMoveRes{
							PlayerId: num,
							ShapeId:  req.ShapeId,
							Pos:      req.Pos,
							Rotation: req.Rotation,
						})
					}
				} else {
				}
			}
			SendMsg(waitingRoom[num].conn, res)
		}
	}
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
	log.Printf("Server listening on %s\n", ln.Addr())

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
