package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"

	squares "github.com/iBug/Squares-go"
)

/* General data format:
 * 2-byte little-endian type
 * 2-byte little-endian length of the following JSON
 * JSON data
 * Req/Res only signifies direction, they do not necessarily correspond
 */

const (
	CONNECT_REQ    = 1
	CONNECT_RES    = 2
	MOVE_REQ       = 3
	MOVE_RES       = 4
	OTHER_MOVE_RES = 5
)

// Connect and retrieve game information
// Also used as a ping
type ConnectReq struct {
	// Empty Id: New connection
	// With Id: Reconnect an existing session
	Id string `json:"id"`
}

type ConnectRes struct {
	Id           string       `json:"id"`        // Empty Id: Auth failure
	PlayerId     int          `json:"player_id"` // Range: 0-3
	Game         squares.Game `json:"game"`
	ActivePlayer int          `json:"active_player"`
}

type MoveReq struct {
	Id       string `json:"id"`
	ShapeId  int    `json:"shape"`
	Pos      [2]int `json:"pos"`
	Rotation int    `json:"rotation"`
}

type MoveRes struct {
	Ok           bool `json:"ok"`
	ActivePlayer int  `json:"active_player"`
}

type OtherMoveRes struct {
	PlayerId int    `json:"player_id"` // Who made the move
	ShapeId  int    `json:"shape"`
	Pos      [2]int `json:"pos"`
	Rotation int    `json:"rotation"`
}

func Marshal(message any) ([]byte, error) {
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
	default:
		return nil, errors.New("not implemented")
	}

	data, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	binary.Write(&b, binary.LittleEndian, msgType)
	binary.Write(&b, binary.LittleEndian, uint16(len(data)))
	b.Write(data)
	return b.Bytes(), nil
}

func Unmarshal(data []byte) (any, error) {
	if len(data) < 4 {
		return nil, errors.New("invalid data")
	}
	msgType := binary.LittleEndian.Uint16(data[:2])
	msgLen := binary.LittleEndian.Uint16(data[2:4])
	if len(data) < int(msgLen)+4 {
		return nil, errors.New("not enough data")
	}

	var message any
	switch msgType {
	case CONNECT_REQ:
		message = ConnectReq{}
	case CONNECT_RES:
		message = ConnectRes{}
	case MOVE_REQ:
		message = MoveReq{}
	case MOVE_RES:
		message = MoveRes{}
	case OTHER_MOVE_RES:
		message = OtherMoveRes{}
	default:
		return nil, errors.New("not implemented")
	}
	err := json.Unmarshal(data[4:msgLen+4], &message)
	return message, err
}
