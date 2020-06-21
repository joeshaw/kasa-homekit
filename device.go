package main

import (
	"encoding/binary"
	"net"
)

type Device struct {
	Addr            net.Addr
	Name            string
	DeviceName      string
	Model           string
	DeviceID        string
	SoftwareVersion string
	RelayState      bool
	OnTime          int
}

func (d *Device) Set(on bool) error {
	const (
		onMsg  = `{"system":{"set_relay_state":{"state":1}}}`
		offMsg = `{"system":{"set_relay_state":{"state":0}}}`
	)

	conn, err := net.Dial("tcp4", d.Addr.String())
	if err != nil {
		return err
	}
	defer conn.Close()

	msg := []byte(offMsg)
	if on {
		msg = []byte(onMsg)
	}

	if err := binary.Write(conn, binary.BigEndian, uint32(len(msg))); err != nil {
		return err
	}

	if _, err := conn.Write(encrypt(msg)); err != nil {
		return err
	}

	buf := make([]byte, 1024)
	_, err = conn.Read(buf)
	return err
}
