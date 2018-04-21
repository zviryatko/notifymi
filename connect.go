package notifymi

import (
	"fmt"
	"encoding/hex"
	"time"

	"github.com/bettercap/gatt"
)

const HANDLE_AUTH_REQUEST = uint16(0x0055)
const HANDLE_CONN_CONTROL = uint16(0x0054)
const RESPONSE_CODE_PAIR = "100101"
const RESPONSE_CODE_RANDOM = "100201"
const RESPONSE_CODE_CONNECTED = "100301"

func sendAuthNotification(p gatt.Peripheral) {
	c := &gatt.Characteristic{}
	c.SetVHandle(HANDLE_AUTH_REQUEST)
	// Pass 0x0100
	b := []byte{1, 0}
	if err := p.WriteCharacteristic(c, b, true); err != nil {
		fmt.Printf("Failed to write characteristic, err: %s\n", err)
	}
}

func sendEncryptionKey(p gatt.Peripheral) {
	c := &gatt.Characteristic{}
	c.SetVHandle(HANDLE_CONN_CONTROL)
	// Pass 0x0100
	b := []byte{1, 0}
	b = append(b, secret.key...)
	if err := p.WriteCharacteristic(c, b, true); err != nil {
		fmt.Printf("Failed to write characteristic, err: %s\n", err)
	}
}

func requestRandomKey(p gatt.Peripheral) {
	c := &gatt.Characteristic{}
	c.SetVHandle(HANDLE_CONN_CONTROL)
	// Pass 0x0200
	b := []byte{2, 0}
	if err := p.WriteCharacteristic(c, b, true); err != nil {
		fmt.Printf("Failed to write characteristic, err: %s\n", err)
	}
}

func confirmPairing(p gatt.Peripheral) {
	c := &gatt.Characteristic{}
	c.SetVHandle(HANDLE_CONN_CONTROL)
	// Pass 0x0100
	b := []byte{3, 0}
	encryptedNumbers, _ := hex.DecodeString(state.random)
	secret.encrypt(encryptedNumbers)
	b = append(b, encryptedNumbers...)
	if err := p.WriteCharacteristic(c, b, true); err != nil {
		fmt.Printf("Failed to write characteristic, err: %s\n", err)
	}
}

func pairDevice(p gatt.Peripheral) {
	// Subscribe to change 0x0054 characteristic.
	c := &gatt.Characteristic{}
	c.SetVHandle(HANDLE_CONN_CONTROL)
	d := &gatt.Descriptor{}
	d.SetHandle(HANDLE_CONN_CONTROL)
	c.SetDescriptor(d)
	p.SetNotifyValue(c, onChangeConnControl)

	sendAuthNotification(p)
	reRunWhenUnlocked(p, state.isPaired, sendEncryptionKey, 3, 10)
	reRunWhenUnlocked(p, state.isRandomNumber, requestRandomKey, 3, 10)
	reRunWhenUnlocked(p, state.isConnected, confirmPairing, 3, 10)
}

func reRunWhenUnlocked(p gatt.Peripheral, check func() bool, run func(p gatt.Peripheral), attemps int, seconds time.Duration) {
	i := 1
	for !check() {
		go run(p)
		i++
		time.Sleep(seconds * time.Second)
		if i > attemps {
			return
		}
	}
}

func onChangeConnControl(c *gatt.Characteristic, data []byte, err error) {
	switch hex.EncodeToString(data[:3]) {
	case RESPONSE_CODE_PAIR:
		// Success pairing
		fmt.Println("Device paired successfully")
		state.Paired()
		break
	case RESPONSE_CODE_RANDOM:
		// Received random key from device, 16 bytes without first 3.
		randomString := hex.EncodeToString(data[3:19])
		state.SetRandomString(randomString)
		fmt.Printf("Received random key: %s\n", randomString)
		break
	case RESPONSE_CODE_CONNECTED:
		// Received random key from device.
		state.Connected()
		fmt.Println("Device is connected!!!")
		break
	}
}
