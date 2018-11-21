package main

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

func (p *program) sendRequest(vhandle uint16, b []byte, per gatt.Peripheral) error {
	c := &gatt.Characteristic{}
	c.SetVHandle(vhandle)
	return per.WriteCharacteristic(c, b, true)
}

func (p *program) sendAuthNotification(per gatt.Peripheral) error {
	fmt.Println("Sending authorization request...")
	if err := p.sendRequest(HANDLE_AUTH_REQUEST, []byte{1, 0}, per); err != nil {
		fmt.Printf("Failed to send auth notification, err: %s\n", err)
		return err
	}
	fmt.Println("Authorization request was sent")
	return p.sendEncryptionKey(per)
}

func (p *program) sendEncryptionKey(per gatt.Peripheral) error {
	message := "Sending encryption key..."
	if err := p.reRunWhenUnlocked(per, p.state.isPaired, append([]byte{1, 0}, p.secret.key...), 3, 10, message); err != nil {
		fmt.Printf("Failed to send encryption key, err: %s\n", err)
		return err
	}
	fmt.Println("Encryption key was sent")
	return p.requestRandomKey(per)
}

func (p *program) requestRandomKey(per gatt.Peripheral) error {
	message := "Requesting random key..."
	if err := p.reRunWhenUnlocked(per, p.state.isRandomNumber, []byte{2, 0}, 3, 10, message); err != nil {
		fmt.Printf("Failed to send encryption key, err: %s\n", err)
		return err
	}
	fmt.Println("Random key received")
	return p.confirmPairing(per)
}

func (p *program) confirmPairing(per gatt.Peripheral) error {
	encryptedNumbers, err := hex.DecodeString(p.state.RandomString())
	if err != nil {
		return err
	}
	p.secret.Encrypt(encryptedNumbers)
	message := "Sending encrypted random key..."
	if err := p.reRunWhenUnlocked(per, p.state.isConnected, append([]byte{3, 0}, encryptedNumbers...), 3, 10, message); err != nil {
		fmt.Printf("Failed to send pairing confirmation, err: %s\n", err)
		return err
	}
	fmt.Println("Encrypted random key was sent")
	return nil
}

func (p *program) pairPeripheral(per gatt.Peripheral) {
	// Subscribe to change 0x0054 characteristic.
	c := &gatt.Characteristic{}
	c.SetVHandle(HANDLE_CONN_CONTROL)
	d := &gatt.Descriptor{}
	d.SetHandle(HANDLE_CONN_CONTROL)
	c.SetDescriptor(d)
	per.SetNotifyValue(c, p.onChangeConnControl)

	if err := p.sendAuthNotification(per); err != nil {
		fmt.Printf("Failed to pair device, err: %s\n", err)
		p.Stop()
	}
}

func (p *program) reRunWhenUnlocked(per gatt.Peripheral, check func() bool, data []byte, attempts int, seconds time.Duration, message string) (err error) {
	for i := 1; i < attempts && !check(); i++ {
		fmt.Println(message)
		err = p.sendRequest(HANDLE_CONN_CONTROL, data, per)
		time.Sleep(seconds * time.Second)
	}
	return err
}

func (p *program) onChangeConnControl(c *gatt.Characteristic, data []byte, err error) {
	code := hex.EncodeToString(data[:3])
	switch code {
	case RESPONSE_CODE_PAIR:
		// Success pairing
		p.state.Paired()
		break
	case RESPONSE_CODE_RANDOM:
		// Received random key from device, 16 bytes without first 3.
		randomString := hex.EncodeToString(data[3:19])
		p.state.SetRandomString(randomString)
		break
	case RESPONSE_CODE_CONNECTED:
		// Received random key from device.
		p.state.Connected()
		break

	default:
		fmt.Printf("Event: %s", code)
	}
}
