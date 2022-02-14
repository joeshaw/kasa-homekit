package main

import (
	"encoding/json"
	"log"
	"net"
	"time"
)

const discoveryMsg = `{"system":{"get_sysinfo":null},"emeter":{"get_realtime":null}}`

type discoveryResponse struct {
	System struct {
		GetSysinfo struct {
			ErrCode int    `json:"err_code"`
			ErrMsg  string `json:"err_msg"`

			SwVer      string `json:"sw_ver"`
			HwVer      string `json:"hw_ver"`
			Model      string `json:"model"`
			DeviceID   string `json:"deviceId"`
			OemID      string `json:"oemId"`
			HwID       string `json:"hwId"`
			Rssi       int    `json:"rssi"`
			LongitudeI int    `json:"longitude_i"`
			LatitudeI  int    `json:"latitude_i"`
			Alias      string `json:"alias"`
			Status     string `json:"status"`
			MicType    string `json:"mic_type"`
			Feature    string `json:"feature"`
			Mac        string `json:"mac"`
			Updating   int    `json:"updating"`
			LedOff     int    `json:"led_off"`
			RelayState int    `json:"relay_state"`
			OnTime     int    `json:"on_time"`
			ActiveMode string `json:"active_mode"`
			IconHash   string `json:"icon_hash"`
			DevName    string `json:"dev_name"`
			NextAction struct {
				Type int `json:"type"`
			} `json:"next_action"`
			Children []struct {
				ID         string `json:"id"`
				State      int    `json:"state"`
				Alias      string `json:"alias"`
				OnTime     int    `json:"on_time"`
				NextAction struct {
					Type int `json:"type"`
				} `json:"next_action"`
			}
			ChildNum int `json:"child_num"`
			NtcState int `json:"ntc_state"`
		} `json:"get_sysinfo"`
	} `json:"system"`
	Emeter struct {
		ErrCode int    `json:"err_code"`
		ErrMsg  string `json:"err_msg"`

		// FIXME: I don't have an energy monitoring plug
	} `json:"emeter"`
}

func discover() ([]Device, error) {
	broadcastAddr, err := net.ResolveUDPAddr("udp4", "255.255.255.255:9999")
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp4", nil)
	if err != nil {
		return nil, err
	}

	_, err = conn.WriteToUDP(encrypt([]byte(discoveryMsg)), broadcastAddr)
	if err != nil {
		return nil, err
	}

	var devices []Device

	for {
		conn.SetReadDeadline(time.Now().Add(1 * time.Second))

		buf := make([]byte, 1024)
		n, raddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				return devices, nil
			}

			return devices, err
		}

		var resp discoveryResponse
		if err := json.Unmarshal(decrypt(buf[:n]), &resp); err != nil {
			log.Printf("Unable to parse discovery JSON response from %s", raddr)
			continue
		}

		si := resp.System.GetSysinfo

		if si.ErrCode != 0 {
			log.Printf("Error response from %s: error code %d, msg: %s", raddr, si.ErrCode, si.ErrMsg)
			continue
		}

		if len(si.Children) == 0 {
			d := Device{
				Addr:            raddr,
				Name:            si.Alias,
				DeviceName:      si.DevName,
				Model:           si.Model,
				DeviceID:        si.DeviceID,
				SoftwareVersion: si.SwVer,
				RelayState:      si.RelayState == 1,
				OnTime:          si.OnTime,
			}
			devices = append(devices, d)
		}

		for c := range si.Children {
			d := Device{
				Addr:            raddr,
				Name:            si.Children[c].Alias,
				DeviceName:      si.DevName,
				Model:           si.Model,
				DeviceID:        si.DeviceID + si.Children[c].ID,
				SoftwareVersion: si.SwVer,
				RelayState:      si.Children[c].State == 1,
				OnTime:          si.Children[c].OnTime,
				IsChild:         true,
			}
			devices = append(devices, d)
		}
	}
}
