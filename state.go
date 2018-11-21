package main

import "fmt"

type State struct {
	paired    bool
	connected bool
	random    string
}

func NewState() *State {
	state := &State{}
	state.paired = false
	state.connected = false
	return state
}

func (state *State) isPaired() bool {
	return state.paired
}

func (state *State) isRandomNumber() bool {
	return len(state.random) > 0
}

func (state *State) isConnected() bool {
	return state.connected
}

func (state *State) Paired() {
	fmt.Println("State changed to: paired")
	state.paired = true
}

func (state *State) Unpaired() {
	fmt.Println("State changed to: unpaired")
	state.paired = false
}

func (state *State) Connected() {
	fmt.Println("State changed to: connected")
	state.connected = true
}

func (state *State) Disconnected() {
	fmt.Println("State changed to: disconnected")
	state.connected = false
}

func (state *State) SetRandomString(random string) {
	fmt.Printf("Received random key: %s\n", random)
	state.random = random
}

func (state *State) RandomString() string {
	return state.random
}