package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
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
	IsChild         bool
}

func (d *Device) Set(on bool) error {
	type system struct {
		SetRelayState struct {
			State int `json:"state"`
		} `json:"set_relay_state"`
	}

	type context struct {
		ChildIDs []string `json:"child_ids"`
	}

	var state struct {
		System  system   `json:"system"`
		Context *context `json:"context,omitempty"`
	}

	if on {
		state.System.SetRelayState.State = 1
	}

	if d.IsChild {
		state.Context = &context{ChildIDs: []string{d.DeviceID}}
	}

	conn, err := net.Dial("tcp4", d.Addr.String())
	if err != nil {
		return err
	}
	defer conn.Close()

	msg, err := json.Marshal(state)
	if err != nil {
		return err
	}

	if err := binary.Write(conn, binary.BigEndian, uint32(len(msg))); err != nil {
		return err
	}

	if _, err := conn.Write(encrypt(msg)); err != nil {
		return err
	}

	var size uint32
	if err := binary.Read(conn, binary.BigEndian, &size); err != nil {
		return err
	}

	buf := make([]byte, size)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}

	var resp struct {
		ErrCode int    `json:"err_code"`
		ErrMsg  string `json:"err_msg"`
	}

	if err := json.Unmarshal(decrypt(buf[:n]), &resp); err != nil {
		return err
	}

	if resp.ErrCode != 0 {
		return fmt.Errorf("%s (err code %d)", resp.ErrMsg, resp.ErrCode)
	}

	return nil
}
