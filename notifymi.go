package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"github.com/bettercap/gatt"
	"github.com/docker/docker-credential-helpers/secretservice"
	"github.com/docker/docker-credential-helpers/credentials"
	"crypto/rand"
	"encoding/hex"
)

var secret *Secret
var state *State
var done = make(chan int)
var nativeStore = secretservice.Secretservice{}

func initSecret(key string) {
	s, err := newSecret(key)
	if err != nil {
		panic(err.Error())
	}
	secret = s
}

func generateCredentials() (string) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err.Error())
	}
	return hex.EncodeToString(b)
}

func initCredentials(deviceId string, regenerate bool) (secret string) {
	_, secret, err := nativeStore.Get(deviceId)
	if regenerate || err != nil {
		secret = generateCredentials()
		c := &credentials.Credentials{
			ServerURL: deviceId,
			Username:  "key",
			Secret:    secret,
		}
		nativeStore.Add(c)
	}
	return
}

func main() {
	flag.Parse()
	if len(flag.Args()) != 1 {
		log.Fatalf("usage: %s [options] peripheral-id\n", os.Args[0])
	}
	devidceId := string(flag.Arg(0))
	initSecret(initCredentials(devidceId, false));
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
