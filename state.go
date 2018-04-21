package main

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
	state.paired = true
}

func (state *State) Unpaired() {
	state.paired = false
}

func (state *State) Connected() {
	state.connected = true
}

func (state *State) Disconnected() {
	state.connected = false
}

func (state *State) SetRandomString(random string) {
	state.random = random
}

func (state *State) RandomString() string {
	return state.random
}