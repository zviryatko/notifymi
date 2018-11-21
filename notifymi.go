package main

import (
	"flag"
	"log"
	"os"
	"github.com/bettercap/gatt"
	"github.com/docker/docker-credential-helpers/secretservice"
	"github.com/docker/docker-credential-helpers/credentials"
	"crypto/rand"
	"encoding/hex"
	"github.com/judwhite/go-svc/svc"
	"fmt"
	"strings"
)

var done = make(chan int)
var nativeStore = secretservice.Secretservice{}

type program struct {
	deviceId string
	device   gatt.Device
	secret   *Secret
	state    *State
	quit     chan struct{}
}

// Initialize connection secret key.
func (p *program) initSecret() (error) {
	key, err := p.getSecretKey(false)
	if err != nil {
		return err
	}
	p.secret, err = NewSecret(key)
	return err
}

// Generate random bytes. Not sure that's completely secure.
func (p *program) generateSalt() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Load secret key from storage or generate new one.
func (p *program) getSecretKey(regenerate bool) (string, error) {
	_, secret, err := nativeStore.Get(p.deviceId)
	if regenerate || err != nil {
		secret, err := p.generateSalt()
		if err != nil {
			return "", err
		}
		c := &credentials.Credentials{
			ServerURL: p.deviceId,
			Username:  "key",
			Secret:    secret,
		}
		err = nativeStore.Add(c)
	}
	return secret, err
}

// Initialize Gatt service with all attached handlers.
func (p *program) initGatt() (err error) {
	options := []gatt.Option{
		gatt.LnxMaxConnections(1),
		gatt.LnxDeviceID(-1, true),
	}

	p.device, err = gatt.NewDevice(options...)
	if err != nil {
		return err
	}
	// todo: Register fsm handlers.
	p.device.Handle(
		gatt.PeripheralDiscovered(p.onPeripheralDiscovered),
		gatt.PeripheralConnected(p.onPeripheralConnected),
		gatt.PeripheralDisconnected(p.onPeripheralDisconnected),
	)
	return
}

func (p *program) onStateChanged(d gatt.Device, s gatt.State) {
	fmt.Println("State:", s)
	switch s {
	case gatt.StateUnknown:
		d.StopScanning()
		return
	case gatt.StateResetting:
		d.StopScanning()
		return
	case gatt.StateUnsupported:
		d.StopScanning()
		return
	case gatt.StateUnauthorized:
		fmt.Println("Pairing...")
		d.StopScanning()
		return
	case gatt.StatePoweredOff:
		d.StopScanning()
		return
	case gatt.StatePoweredOn:
		fmt.Println("Scanning...")
		d.Scan([]gatt.UUID{}, true)
		return
	default:
		d.StopScanning()
	}
}

func (p *program) onPeripheralDiscovered(per gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	id := strings.ToUpper(flag.Args()[0])

	if strings.ToUpper(per.ID()) != id {
		return
	}

	// Stop scanning once we've got the peripheral we're looking for.
	p.device.StopScanning()

	fmt.Println("  Peripheral ID:", per.ID(), "NAME:(", per.Name(), ")")
	fmt.Println("  Local Name        =", a.LocalName)
	fmt.Println("  TX Power Level    =", a.TxPowerLevel)
	fmt.Println("  Manufacturer Data =", a.ManufacturerData)
	fmt.Println("  Service Data      =", a.ServiceData)
	fmt.Println("")

	if !p.state.isConnected() {
		p.device.Connect(per)
	}
}

func (p *program) onPeripheralConnected(per gatt.Peripheral, err error) {
	fmt.Println("Connected. Pairing...")
	//defer p.Device().CancelConnection(p)

	if err := per.SetMTU(500); err != nil {
		fmt.Printf("Failed to set MTU, err: %s\n", err)
	}
	go p.pairPeripheral(per)

	// Discovery services
	//ss, err := p.DiscoverServices(nil)
	//if err != nil {
	//	fmt.Printf("Failed to discover services, err: %s\n", err)
	//	return
	//}
	//
	//for _, s := range ss {
	//	msg := "Service: " + s.UUID().String()
	//	if len(s.Name()) > 0 {
	//		msg += " (" + s.Name() + ")"
	//	}
	//	fmt.Println(msg)
	//
	//	// Discovery characteristics
	//	cs, err := p.DiscoverCharacteristics(nil, s)
	//	if err != nil {
	//		fmt.Printf("Failed to discover characteristics, err: %s\n", err)
	//		continue
	//	}
	//
	//	for _, c := range cs {
	//		msg := "  Characteristic  " + c.UUID().String()
	//		if len(c.Name()) > 0 {
	//			msg += " (" + c.Name() + ")"
	//		}
	//		msg += "\n    properties    " + c.Properties().String()
	//		fmt.Println(msg)
	//
	//		// Read the characteristic, if possible.
	//		if (c.Properties() & gatt.CharRead) != 0 {
	//			b, err := p.ReadCharacteristic(c)
	//			if err != nil {
	//				fmt.Printf("Failed to read characteristic, err: %s\n", err)
	//				continue
	//			}
	//			fmt.Printf("    value         %x | %q\n", b, b)
	//		}
	//
	//		// Discovery descriptors
	//		ds, err := p.DiscoverDescriptors(nil, c)
	//		if err != nil {
	//			fmt.Printf("Failed to discover descriptors, err: %s\n", err)
	//			continue
	//		}
	//
	//		for _, d := range ds {
	//			msg := "  Descriptor      " + d.UUID().String()
	//			if len(d.Name()) > 0 {
	//				msg += " (" + d.Name() + ")"
	//			}
	//			fmt.Println(msg)
	//
	//			// Read descriptor (could fail, if it's not readable)
	//			b, err := p.ReadDescriptor(d)
	//			if err != nil {
	//				fmt.Printf("Failed to read descriptor, err: %s\n", err)
	//				continue
	//			}
	//			fmt.Printf("    value         %x | %q\n", b, b)
	//		}
	//
	//		// Subscribe the characteristic, if possible.
	//		if (c.Properties() & (gatt.CharNotify | gatt.CharIndicate)) != 0 {
	//			f := func(c *gatt.Characteristic, b []byte, err error) {
	//				fmt.Printf("notified: % X | %q\n", b, b)
	//			}
	//			if err := p.SetNotifyValue(c, f); err != nil {
	//				fmt.Printf("Failed to subscribe characteristic, err: %s\n", err)
	//				continue
	//			}
	//		}
	//
	//	}
	//	fmt.Println()
	//}
}

func (p *program) onPeripheralDisconnected(per gatt.Peripheral, err error) {
	close(done)
}

func (p *program) Init(env svc.Environment) (error) {
	p.state = NewState()
	if err := p.initSecret(); err != nil {
		return err
	}
	return p.initGatt()
}

// Start the service.
func (p *program) Start() (error) {
	// The Start method must not block, or Windows may assume your service failed
	// to start. Launch a Goroutine here to do something interesting/blocking.

	p.quit = make(chan struct{})
	go p.device.Init(p.onStateChanged)
	return nil
}

// Stop the service.
func (p *program) Stop() (error) {
	// The Stop method is invoked by stopping the Windows service, or by pressing Ctrl+C on the console.
	// This method may block, but it's a good idea to finish quickly or your process may be killed by
	// Windows during a shutdown/reboot. As a general rule you shouldn't rely on graceful shutdown.

	close(p.quit)
	return p.device.Stop()
}

func main() {
	flag.Parse()
	if len(flag.Args()) != 1 {
		log.Fatalf("Usage: %s [options] peripheral-id\n", os.Args[0])
	}
	prg := &program{deviceId: string(flag.Arg(0))}
	if err := svc.Run(prg); err != nil {
		log.Fatal(err)
	}
}
