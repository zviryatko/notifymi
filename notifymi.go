package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"github.com/bettercap/gatt"
)

var secret *Secret
var state *State
var done = make(chan int)

func initSecret() {
	// Todo: move somewhere to safe place.
	key := "5d19ebd2014c847eb1621a9baa625bd3"

	s, err := newSecret(key)
	if err != nil {
		panic(err.Error())
	}
	secret = s
}

func main() {
	flag.Parse()
	if len(flag.Args()) != 1 {
		log.Fatalf("usage: %s [options] peripheral-id\n", os.Args[0])
	}

	initSecret()
	state = NewState()

	options := []gatt.Option{
		gatt.LnxMaxConnections(1),
		gatt.LnxDeviceID(-1, true),
	}

	d, err := gatt.NewDevice(options...)
	if err != nil {
		log.Fatalf("Failed to open device, err: %s\n", err)
		return
	}

	// Register handlers.
	d.Handle(
		gatt.PeripheralDiscovered(onPeriphDiscovered),
		gatt.PeripheralConnected(onPeriphConnected),
		gatt.PeripheralDisconnected(onPeriphDisconnected),
	)

	d.Init(onStateChanged)
	fmt.Println("Done")
	<-done
	os.Exit(0)
}
