package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"time"
)

const (
	// Micro Default IDs

	MicroIDNoResponse = MicroID(-1)
	MicroIDDefault    = MicroID(0)

	// Micro Types

	MicroTypeGET = MicroType("get")
	MicroTypeSET = MicroType("set")

	// Micro Groups

	MicroGroupConfig = MicroGroup("config")
	MicroGroupLED    = MicroGroup("led")

	// MicroGroupInfo only used for `MicroTypeGET`
	MicroGroupInfo = MicroGroup("led")
)

type (
	MicroID    int
	MicroType  string
	MicroGroup string
)

type MicroRequest struct {
	MicroSocket

	ID    MicroID    `json:"id"`
	Group MicroGroup `json:"group"`
	Type  MicroType  `json:"type"`

	// Command mappings:
	//
	// Type: "set":
	// 		- Group: "config" Command: "pins"
	// 		- Group: "led"    Command: "color"
	//
	// Type: "get":
	// 		- Group: "config" Command: "pins"
	// 		- Group: "led"    Command: "color"
	// 		- Group: "info"   Command: "temp"
	// 		- Group: "info"   Command: "disk-usage"
	// 		- Group: "info"   Command: "version"
	Command string `json:"command"`

	// CommandArgs can be nil
	//
	// 	if type is `MicroTypeSET` and group is `MicroGroupConfig`
	//		if command is "pins":
	// 			[]uint8 - range between 0-28 converted to a slice with strings
	//					  https://i0.wp.com/randomnerdtutorials.com/wp-content/uploads/2024/02/Raspberry-Pi-Pico-W-RP2040-Rev3-Board-Pinout-GPIOs.png?quality=100&strip=all&ssl=1
	//
	// 	elif type is `MicroTypeSET` and group is `MicroGroupLED`
	//		if command is "color":
	// 			[]uint8 - range between 0-255 converted to a slice with strings
	//
	// 	else
	// 		nil
	CommandArgs []string `json:"args"`
}

func NewMicroRequest(id MicroID) *MicroRequest {
	return &MicroRequest{
		ID: id,
	}
}

func (mr *MicroRequest) Send(d *Device) ([]byte, error) {
	d.Online = true
	d.Error = ""

	if !mr.IsConnected() {
		if err := mr.Connect(d.Server.Addr); err != nil {
			d.Online = false
			d.Error = err.Error()
			return nil, err
		}
		// NOTE: Ok, all connections created here will be closed, all other
		// 		 connections will stay open
		defer mr.Close()
	}

	if mr.CommandArgs == nil {
		mr.CommandArgs = []string{}
	}

	data, err := json.Marshal(mr)
	if err != nil {
		panic(err.Error())
	}
	err = mr.Write(data)
	if err != nil {
		mr.socket = nil
		d.Error = err.Error()
		d.Online = false
		return nil, err
	}

	if mr.ID == MicroIDNoResponse {
		return []byte{}, nil
	}

	data, err = mr.Read()
	if err != nil {
		mr.socket = nil
		d.Error = err.Error()
		d.Online = false
		return nil, err
	}

	slog.Debug("Got some data",
		"device.name", d.Server.Name, "device.address", d.Server.Addr,
		"command", fmt.Sprintf("(%d) %s %s %s %+v",
			mr.ID, mr.Type, mr.Group, mr.Command, mr.CommandArgs,
		),
	)

	return data, nil
}

// RequestPins will change fields like "ID", "Type", "Group", "Command" or
// "CommandArgs"
func (mr *MicroRequest) Pins(d *Device) (MicroPins, error) {
	mr.Type = MicroTypeGET
	mr.Group = MicroGroupConfig
	mr.Command = "pins"
	mr.CommandArgs = []string{}

	data, err := mr.Send(d)
	if err != nil {
		return nil, err
	}
	if d.Error != "" {
		return nil, errors.New(d.Error)
	}

	pins, err := ParseMicroResponse[MicroPins](data)
	if err != nil {
		d.Error = err.Error()
	}
	return pins, err
}

func (mr *MicroRequest) Color(d *Device) (MicroColor, error) {
	mr.Type = MicroTypeGET
	mr.Group = MicroGroupLED
	mr.Command = "color"
	mr.CommandArgs = []string{}

	data, err := mr.Send(d)
	if err != nil {
		return nil, err
	}
	if d.Error != "" {
		return nil, errors.New(d.Error)
	}

	color, err := ParseMicroResponse[MicroColor](data)
	if err != nil {
		d.Error = err.Error()
	}
	return color, err
}

func (mr *MicroRequest) SetColor(d *Device, c MicroColor) error {
	mr.Type = MicroTypeSET
	mr.Group = MicroGroupLED
	mr.Command = "color"
	mr.CommandArgs = []string{}
	for _, n := range c {
		mr.CommandArgs = append(mr.CommandArgs, strconv.Itoa(int(n)))
	}

	data, err := mr.Send(d)
	if err != nil {
		return err
	}
	if d.Error != "" {
		return errors.New(d.Error)
	}

	_, err = ParseMicroResponse[any](data)
	if err != nil {
		d.Error = err.Error()
	}
	return err
}

func (mr *MicroRequest) SetPins(d *Device, p MicroPins) error {
	mr.Type = MicroTypeSET
	mr.Group = MicroGroupConfig
	mr.Command = "pins"
	mr.CommandArgs = []string{}
	for _, n := range p {
		mr.CommandArgs = append(mr.CommandArgs, strconv.Itoa(int(n)))
	}

	data, err := mr.Send(d)
	if err != nil {
		return err
	}
	if d.Error != "" {
		return errors.New(d.Error)
	}

	_, err = ParseMicroResponse[any](data)
	if err != nil {
		d.Error = err.Error()
	}
	return err
}

type (
	MicroPins      []uint
	MicroColor     []uint
	MicroTemp      int
	MicroDiskUsage struct {
		Used int `json:"used"`
		Free int `json:"free"`
	}
	MicroVersion string
)

type MicroSocket struct {
	socket net.Conn
}

func (ms *MicroSocket) IsConnected() bool {
	return ms.socket != nil
}

func (ms *MicroSocket) Connect(addr string) error {
	if ms.socket != nil {
		ms.Close()
	}

	dialer := net.Dialer{
		Timeout: time.Duration(time.Second * 5),
	}

	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		ms.socket = nil
	} else {
		ms.socket = conn
	}

	return err
}

func (ms *MicroSocket) Write(data []byte) error {
	data = append(data, []byte("\n")...)
	n, err := ms.socket.Write(data)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("write: no data")
	}

	return nil
}

func (ms *MicroSocket) Read() ([]byte, error) {
	if ms.socket == nil {
		panic("socket is nil, call connect first")
	}

	data := make([]byte, 0)
	b := make([]byte, 1)
	endByte := []byte("\n")
	var n int
	var err error
	for {
		n, err = ms.socket.Read(b)
		if err != nil {
			break
		}
		if n == 0 {
			err = errors.New("read: no data")
			break
		}

		if bytes.Equal(b, endByte) {
			break
		}

		data = append(data, b...)
	}

	return data, err
}

func (ms *MicroSocket) Close() {
	if ms.socket == nil {
		return
	}

	ms.socket.Close()
	ms.socket = nil
}

type MicroResponse[T any | MicroPins | MicroColor | MicroTemp | MicroDiskUsage | MicroVersion] struct {
	ID    MicroID `json:"id"`
	Error string  `json:"error"`

	// Data contains the data requested
	Data T `json:"data"`
}

func ParseMicroResponse[T any](data []byte) (T, error) {
	resp := &MicroResponse[T]{}
	err := json.Unmarshal(data, resp)
	if err != nil {
		return resp.Data, err
	}
	if resp.Error != "" {
		return resp.Data, errors.New(resp.Error)
	}
	return resp.Data, nil
}
