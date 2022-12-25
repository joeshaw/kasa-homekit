package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	hclog "github.com/brutella/hc/log"
)

const (
	discoveryInterval = 5 * time.Second
)

var devicePath = filepath.Join(os.Getenv("HOME"), ".homecontrol", "kasa")

type kasaAccessory struct {
	device    Device
	acc       *accessory.Switch
	transport hc.Transport
	lastSeen  time.Time
}

type kasaAccessoryMap struct {
	m  map[string]*kasaAccessory
	mu sync.Mutex
}

// ExpireOld cleans up devices that haven't been seen in a while
func (kam *kasaAccessoryMap) ExpireOld() {
	kam.mu.Lock()
	defer kam.mu.Unlock()

	if kam.m == nil {
		return
	}

	for _, ka := range kam.m {
		if time.Since(ka.lastSeen) > 5*time.Minute {
			log.Printf("Dropping lost accessory %q", ka.device.Name)
			<-ka.transport.Stop()
			delete(kam.m, ka.device.DeviceID)
		}
	}
}

func (kam *kasaAccessoryMap) AddOrUpdate(d Device) error {
	kam.mu.Lock()
	defer kam.mu.Unlock()

	if kam.m == nil {
		kam.m = map[string]*kasaAccessory{}
	}

	if a, ok := kam.m[d.DeviceID]; ok {
		a.device = d
		a.acc.Switch.On.SetValue(d.RelayState)
		a.lastSeen = time.Now()
		return nil
	}

	info := accessory.Info{
		Name:             d.Name,
		Model:            d.Model,
		Manufacturer:     "TP-Link",
		FirmwareRevision: d.SoftwareVersion,
		SerialNumber:     d.DeviceID,
	}
	acc := accessory.NewSwitch(info)
	acc.Switch.On.SetValue(d.RelayState)

	hcConfig := hc.Config{
		Pin:         "00102003",
		StoragePath: filepath.Join(devicePath, d.DeviceID),
	}

	t, err := hc.NewIPTransport(hcConfig, acc.Accessory)
	if err != nil {
		return fmt.Errorf("Unable to create homekit device: %w", err)
	}

	acc.Switch.On.OnValueRemoteUpdate(func(on bool) {
		if err := d.Set(on); err != nil {
			log.Printf("Unable to update switch %q to %t: %v", d.Name, on, err)
		}
	})

	log.Printf("Creating accessory for %q at %v", d.Name, d.Addr)

	ka := &kasaAccessory{
		device:    d,
		acc:       acc,
		transport: t,
		lastSeen:  time.Now(),
	}

	go t.Start()

	kam.m[d.DeviceID] = ka
	return nil
}

func main() {
	if x := os.Getenv("DEBUG"); x != "" {
		hclog.Debug.Enable()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(ctx context.Context) {
		log.Printf("Starting device discovery loop")
		defer log.Printf("Exiting device discovery loop")

		var kam kasaAccessoryMap

		// The first loop iteration runs immediately, subsequent ones
		// run on discoveryInterval.
		var interval time.Duration

		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(interval):
				devices, err := discover()
				if err != nil {
					log.Printf("error discovering devices: %v", err)
					break
				}

				for _, d := range devices {
					if err := kam.AddOrUpdate(d); err != nil {
						log.Println(err)
					}
				}

			}

			interval = discoveryInterval
		}
	}(ctx)

	hc.OnTermination(func() {
		cancel()
	})

	<-ctx.Done()
}
