package midi

/*
These tests require IAC buses to be created on an OS X system, named:
    Bus 1
    Bus 2
    Bus 3
*/

import (
	"testing"
	"time"
)

// TODO: Rename Event to Message since there's no time data.
func TestPipe(t *testing.T) {
	devices, _ := GetDevices()
	iac1, _ := devices["IAC Driver Bus 1"]
	iac2, _ := devices["IAC Driver Bus 2"]
	pipe, _ := NewPipe(iac1, iac2)
	go pipe.Connect()
	expected := NoteOn{0, 64, 127}
	// Spoof a MIDI note coming into the device.
	pipe.From.OutPort().Events() <- expected
	actual := <-pipe.To.OutPort().Events()
	if expected != actual {
		t.Errorf("Received %q from pipe instead of %q", actual, expected)
	}
	pipe.Stop()
	devices.Shutdown()
}

// TODO: This test crashes out sometimes. Why? (PortMidi init times?)
func testChain(t *testing.T) {
	devices, _ := GetDevices()
	iac1, _ := devices["IAC Driver Bus 1"]
	iac2, _ := devices["IAC Driver Bus 2"]
	iac3, _ := devices["IAC Driver Bus 3"]

	chain, _ := NewChain(iac1, iac2, iac3)
	go chain.Connect()

	expected := NoteOn{0, 64, 127}
	chain.Devices[0].OutPort().Events() <- expected
	actual := <-chain.Devices[2].OutPort().Events()

	if expected != actual {
		t.Errorf("Received %q from chain instead of %q", actual, expected)
	}
	chain.Stop()
	devices.Shutdown()
}

func TestRouter(t *testing.T) {
	devices, _ := GetDevices()
	iac1 := devices["IAC Driver Bus 1"]
	iac2 := devices["IAC Driver Bus 2"]
	iac3 := devices["IAC Driver Bus 3"]
	router, _ := NewRouter(iac1, iac2, iac3)
	go router.Connect()
	expected := NoteOn{0, 64, 127}
	router.From.OutPort().Events() <- expected
	actual1 := <-router.To[0].OutPort().Events()
	actual2 := <-router.To[1].OutPort().Events()
	if expected != actual1 || expected != actual2 {
		t.Errorf("Recived %q and %q from router instead of %q",
			actual1, actual2, expected)
	}
	router.Stop()
	devices.Shutdown()
}

// TODO: This test crashes out sometimes. Why? (PortMidi init times?)
func testFunnel(t *testing.T) {
	devices, _ := GetDevices()
	iac1 := devices["IAC Driver Bus 1"]
	iac2 := devices["IAC Driver Bus 2"]
	iac3 := devices["IAC Driver Bus 3"]
	funnel, _ := NewFunnel(iac1, iac2, iac3)
	go funnel.Connect()
	expected := NoteOn{0, 64, 127}
	funnel.From[1].OutPort().Events() <- expected
	actual := <-funnel.To.OutPort().Events()
	if expected != actual {
		t.Errorf("Received %q from funnel instead of %q",
			actual, expected)
	}
	expected = NoteOn{0, 95, 64}
	funnel.From[0].OutPort().Events() <- expected
	actual = <-funnel.To.OutPort().Events()
	if expected != actual {
		t.Errorf("Received %q from funnel instead of %q",
			actual, expected)
	}
	funnel.Stop()
	devices.Shutdown()
}

func TestSystemDevice(t *testing.T) {
	devices, _ := GetDevices()
	iac1, _ := devices["IAC Driver Bus 1"]
	iac1.Open()
	iac1.Run()
	iac1.Close()
	devices.Shutdown()
}

func TestThruDevice(t *testing.T) {
	thru := NewThruDevice()
	thru.Open()
	go thru.Run()
	expected := NoteOn{0, 64, 127}
	thru.InPort().Events() <- expected
	actual := <-thru.OutPort().Events()
	if expected != actual {
		t.Errorf("Received %q from ThruDevice instead of %q", actual, expected)
	}
}

func ExamplePipe() {
	devices, _ := GetDevices()
	nanoPad := devices["nanoPAD2 PAD"]
	iac1 := devices["IAC Driver Bus 1"]
	pipe, _ := NewPipe(nanoPad, iac1)
	go pipe.Connect()
	time.Sleep(5 * time.Second)
	pipe.Stop()
	devices.Shutdown()
}

func ExampleRouter() {
	devices, _ := GetDevices()
	nanoPad := devices["nanoPAD2 PAD"]
	iac1 := devices["IAC Driver Bus 1"]
	iac2 := devices["IAC Driver Bus 2"]
	router, _ := NewRouter(nanoPad, iac1, iac2)
	go router.Connect()
	time.Sleep(5 * time.Second)
	router.Stop()
	devices.Shutdown()
}

func ExampleChain() {
	devices, _ := GetDevices()
	nanoPad, _ := devices["nanoPAD2 PAD"]
	iac1, _ := devices["IAC Driver Bus 1"]
	iac2, _ := devices["IAC Driver Bus 2"]
	chain, _ := NewChain(nanoPad, iac1, iac2)
	go chain.Connect()
	time.Sleep(1 * time.Minute)
	chain.Stop()
	devices.Shutdown()
}

func ExampleTransposer() {
	devices, _ := GetDevices()
	nanoPad := devices["nanoPAD2 PAD"]
	transposer := NewTransposer(map[int]int{36: 37, 37: 36}, nil)
	iac1 := devices["IAC Driver Bus 1"]
	chain, _ := NewChain(nanoPad, transposer, iac1)
	go chain.Connect()
	time.Sleep(1 * time.Minute)
	chain.Stop()
	devices.Shutdown()
}

func ExampleChannelTransposer() {
	// For use with midi_fractals.pde
	devices, _ := GetDevices()
	iac1 := devices["IAC Driver Bus 1"]
	iac2 := devices["IAC Driver Bus 2"]
	transposer := NewTransposer(
		map[int]int{1: 36, 2: 37, 3: 38, 4: 40, 5: 41, 6: 42},
		func(t Transposer) {
			for {
				select {
				case note := <-t.InPort().Events():
					if key, ok := t.NoteMap[note.Channel]; ok {
						note.Channel = 0
						note.Key = key
						t.OutPort().Events() <- note
					}
				case note := <-t.InPort().Events():
					if key, ok := t.NoteMap[note.Channel]; ok {
						note.Channel = 0
						note.Key = key
						t.OutPort().Events() <- note
					}
				}
			}
		})
	chain, _ := NewChain(iac1, transposer, iac2)
	go chain.Connect()
	c := make(chan int)
	<-c // Block forever
	chain.Stop()
	devices.Shutdown()
}

func ExampleNanopad() {
	devices, _ := GetDevices()

	nanopad := devices["nanoPAD PAD"]
	nanopad2 := devices["nanoPAD2 PAD"]
	iac1 := devices["IAC Driver Bus 1"]

	// Make top row of nanopad 1 have similar button mapping to nanopad 2.
	trans := NewTransposer(
		map[int]int{39: 37, 48: 39, 45: 41, 51: 45, 49: 47}, nil)

	chain, _ := NewChain(nanopad, trans, iac1)
	go chain.Connect()

	pipe, _ := NewPipe(nanopad2, iac1)
	go pipe.Connect()

	select {}
}
