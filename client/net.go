package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	squares "github.com/iBug/Squares-go"
)

/* General data format:
 * 2-byte little-endian type
 * 2-byte little-endian length of the following JSON
 * JSON data
 * Req/Res only signifies direction, they do not necessarily correspond
 */

// Using iota: https://go.dev/ref/spec#Constant_declarations
const (
	_ = iota
	CONNECT_REQ
	CONNECT_RES
	MOVE_REQ
	MOVE_RES
	OTHER_MOVE_RES
	SERVER_RES // Generic server message
)

const (
	// Server responses
	_ = iota
	S_CLIENT_REJECTED
	S_GAME_NOT_GOING
	S_GAME_OVER
)

// Description of server messages
var SERVER_RES_S = map[int]string{
	S_CLIENT_REJECTED: "client rejected",
	S_GAME_NOT_GOING:  "game not going",
	S_GAME_OVER:       "game is over",
}

func ServerResString(i int) string {
	s, ok := SERVER_RES_S[i]
	if !ok {
		return fmt.Sprintf("unknown server message %d", i)
	}
	return s
}

// Connect and retrieve game information
// Also used as a ping
type ConnectReq struct {
	// Empty Id: New connection
	// With Id: Reconnect an existing session
	Id int `json:"id"`
}

type ConnectRes struct {
	Id       int          `json:"id"`        // Empty Id: Auth failure
	PlayerId int          `json:"player_id"` // Range: 0-3
	Game     squares.Game `json:"game"`
}

type MoveReq struct {
	Id       int    `json:"id"`
	ShapeId  int    `json:"shape"`
	Pos      [2]int `json:"pos"`
	Rotation int    `json:"rotation"`
}

type MoveRes struct {
	Ok           bool `json:"ok"`
	ActivePlayer int  `json:"active_player"`
}

type OtherMoveRes struct {
	PlayerId     int    `json:"player_id"` // Who made the move
	ShapeId      int    `json:"shape"`
	Pos          [2]int `json:"pos"`
	Rotation     int    `json:"rotation"`
	ActivePlayer int    `json:"active_player"`
}

type ServerRes struct {
	Code int `json:"code"`
}

func SendMsg(w io.Writer, message any) error {
	var msgType uint16
	switch message.(type) {
	case ConnectReq:
		msgType = CONNECT_REQ
	case ConnectRes:
		msgType = CONNECT_RES
	case MoveReq:
		msgType = MOVE_REQ
	case MoveRes:
		msgType = MOVE_RES
	case OtherMoveRes:
		msgType = OTHER_MOVE_RES
	case ServerRes:
		msgType = SERVER_RES
	default:
		return errors.New("not implemented")
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	binary.Write(w, binary.LittleEndian, msgType)
	binary.Write(w, binary.LittleEndian, uint16(len(data)))
	_, err = w.Write(data)
	return err
}

func RecvMsg(r io.Reader) (any, error) {
	b := make([]byte, 2)
	_, err := io.ReadFull(r, b)
	if err != nil {
		return nil, err
	}
	msgType := binary.LittleEndian.Uint16(b)
	_, err = io.ReadFull(r, b)
	if err != nil {
		return nil, err
	}
	msgLen := binary.LittleEndian.Uint16(b)

	data := make([]byte, msgLen)
	_, err = io.ReadFull(r, data)
	if err != nil {
		return nil, err
	}

	var message any
	switch msgType {
	case CONNECT_REQ:
		m := ConnectReq{}
		err = json.Unmarshal(data, &m)
		message = m
	case CONNECT_RES:
		m := ConnectRes{}
		err = json.Unmarshal(data, &m)
		message = m
	case MOVE_REQ:
		m := MoveReq{}
		err = json.Unmarshal(data, &m)
		message = m
	case MOVE_RES:
		m := MoveRes{}
		err = json.Unmarshal(data, &m)
		message = m
	case OTHER_MOVE_RES:
		m := OtherMoveRes{}
		err = json.Unmarshal(data, &m)
		message = m
	case SERVER_RES:
		m := ServerRes{}
		err = json.Unmarshal(data, &m)
		message = m
	default:
		return nil, fmt.Errorf("not implemented: %d", msgType)
	}
	return message, err
}
