package main

import (
	"errors"
	"io"
	"log"
	"math/rand"
	"net"
	"time"

	squares "github.com/iBug/Squares-go"
)

// Use as "connection control block"
type ClientInfo struct {
	id   int
	conn net.Conn
}

type ClientMessage struct {
	ci *ClientInfo
	m  any
}

type ClientDisconnect struct{} // Internal data type

const IDRANGE = 999999999

var (
	lobby       []*ClientInfo
	gameOngoing = false

	r1 = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func generateClientID() int {
	return r1.Intn(IDRANGE) + 1
}

func findClientInfo(ci *ClientInfo) int {
	for i := range lobby {
		if lobby[i] == ci {
			return i
		}
	}
	return -1
}

func getAvailableLobbySlot() int {
	for i := range lobby {
		if lobby[i] == nil {
			return i
		}
	}
	lobby = append(lobby, nil)
	return len(lobby) - 1
}

func handleClient(ci *ClientInfo, ch chan<- ClientMessage) {
	defer ci.conn.Close()
	for {
		msg, err := RecvMsg(ci.conn)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				// serverMain closed the connection
				// pass
			} else if err == io.EOF {
				log.Printf("Client %d disconnected\n", ci.id)
			} else {
				log.Println(err)
			}
			ch <- ClientMessage{ci, ClientDisconnect{}}
			break
		}
		ch <- ClientMessage{ci, msg}
	}
}

func processClientMessage(cm ClientMessage) {
	ci := cm.ci
	num := findClientInfo(ci)
SwitchCMType:
	switch req := cm.m.(type) {
	case ConnectReq:
		if ci.id == 0 && req.Id != 0 {
			// Register ID with connection control block
			ci.id = req.Id
		}
		if num != -1 {
			// Existing connection as ping
			SendMsg(ci.conn, ConnectRes{ci.id, num, *game})
			break
		}

		if gameOngoing {
			// Handle potential reconnection
			for i := range lobby {
				if req.Id == lobby[i].id {
					log.Printf("Client[%d] %d reconnected as %s (was %s)\n",
						i, ci.id, ci.conn.RemoteAddr(), lobby[i].conn.RemoteAddr())
					lobby[i].conn.Close()
					lobby[i] = ci
					SendMsg(ci.conn, ConnectRes{ci.id, i, *game})
					break SwitchCMType
				}
			}
			// Unrecognized connection
			log.Printf("Client[?] %d connected while game ongoing\n", ci.id)
			SendMsg(ci.conn, ServerRes{S_UNKNOWN})
			ci.conn.Close()
		} else {
			// New connection as join request
			ci.id = req.Id
			if ci.id == 0 {
				ci.id = generateClientID()
			}
			num = len(lobby)
			slot := getAvailableLobbySlot()
			lobby[slot] = ci
			if len(lobby) == squares.NPLAYERS {
				// Start game
				gameOngoing = true
				game.Reset()
			}
			SendMsg(ci.conn, ConnectRes{ci.id, num, *game})
		}
	case MoveReq:
		if !gameOngoing {
			SendMsg(ci.conn, ServerRes{S_GAME_NOT_GOING})
			break
		}
		if num != game.ActivePlayer {
			log.Printf("Received move from inactive player %d\n", num)
			break
		}

		pos := squares.Coord{req.Pos[0], req.Pos[1]}
		if !game.TryInsert(req.ShapeId, req.Rotation, pos, num, game.FirstRound) {
			SendMsg(ci.conn, MoveRes{
				Ok:           false,
				ActivePlayer: game.ActivePlayer,
			})
			break
		}
		game.Insert(req.ShapeId, req.Rotation, pos, num)
		game.AfterMove()
		for i := 0; i < squares.NPLAYERS; i++ {
			SendMsg(lobby[i].conn, OtherMoveRes{
				PlayerId:     num,
				ShapeId:      req.ShapeId,
				Pos:          req.Pos,
				Rotation:     req.Rotation,
				ActivePlayer: game.ActivePlayer,
			})
		}
	case ClientDisconnect:
		if num == -1 {
			break
		}
		if gameOngoing {
			log.Printf("Client[%d] %d disconnected while game ongoing", num, ci.id)
			// just wait for reconnection
		} else {
			lobby[num] = nil
		}
	default:
		log.Printf("Unknown message type: %T\n", req)
	}
}

func serverGame(chCM <-chan ClientMessage) {
	for {
		processClientMessage(<-chCM)
	}
}

func resetLobby() {
	lobby = make([]*ClientInfo, 0, 2*squares.NPLAYERS)
}

func serverMain() {
	resetLobby()

	ln, err := net.Listen("tcp", fServerAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	log.Printf("Server listening on %s\n", ln.Addr())

	chCI := make(chan *ClientInfo, 1)
	chCM := make(chan ClientMessage, 8)
	go serverGame(chCM)
	defer close(chCI)
	defer close(chCM)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			conn.Close()
			continue
		}
		go handleClient(&ClientInfo{conn: conn}, chCM)
	}
}
