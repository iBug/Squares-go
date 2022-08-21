package main

import (
	"log"
	"math/rand"
	"net"
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
	waitingRoom []*ClientInfo
	gamePlayers [4]*ClientInfo
	gameOngoing = false

	r1 = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func handleClient(num int, ci *ClientInfo, ch chan<- ClientMessage) {
	defer ci.conn.Close()
	defer close(ch)
	for {
		msg, err := RecvMsg(ci.conn)
		if err != nil {
			log.Println(err)
			break
		}
		ch <- ClientMessage{num, msg}
	}
}

func serverGame(chCI <-chan *ClientInfo) {
	chCM := make(chan ClientMessage)

	// GameLoop:
	for {
		select {
		case ci := <-chCI:
			if gameOngoing {
				log.Printf("Client %d connected while game ongoing\n", ci.id)
				ci.conn.Close()
				break
			}
			wrCount := len(waitingRoom)
			waitingRoom = append(waitingRoom, ci)
			go handleClient(wrCount, waitingRoom[wrCount], chCM)
			wrCount++
			if wrCount == squares.NPLAYERS {
				gameOngoing = true
				copy(gamePlayers[:], waitingRoom[:4])
				game.Reset()
			}
		case cm := <-chCM:
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
				if !gameOngoing {
					break
				}
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
						SendMsg(waitingRoom[i].conn, OtherMoveRes{
							PlayerId:     num,
							ShapeId:      req.ShapeId,
							Pos:          req.Pos,
							Rotation:     req.Rotation,
							ActivePlayer: game.ActivePlayer,
						})
					}
				}
			default:
				log.Printf("Unknown message type: %T\n", req)
			}
			SendMsg(waitingRoom[num].conn, res)
		}
	}
}

func resetWaitingRoom() {
	waitingRoom = make([]*ClientInfo, 0, 2*squares.NPLAYERS)
}

func serverMain() {
	resetWaitingRoom()

	ln, err := net.Listen("tcp", fServerAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	log.Printf("Server listening on %s\n", ln.Addr())

	ch := make(chan *ClientInfo, 1)
	go serverGame(ch)
	defer close(ch)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			conn.Close()
			continue
		}
		ch <- &ClientInfo{conn: conn}
	}
}
