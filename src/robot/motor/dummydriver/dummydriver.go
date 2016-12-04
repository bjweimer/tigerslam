// Package dummydriver was created because of lack of availability of a motor
// driver card when testing easing packages. It should confirm the working of
// easing mechanisms in some way. It implements the driver interface.
//
// The current implementation will just print the speeds which are set.
package dummydriver

import (
	"fmt"
)

type DummyDriver struct {
	connected bool
}

func (d *DummyDriver) Connect() error {
	d.connected = true
	return nil
}

func (d *DummyDriver) Disconnect() error {
	d.connected = false
	return nil
}

func (d *DummyDriver) IsConnected() bool {
	return d.connected
}

func (d *DummyDriver) SetSpeeds(left, right float64) error {
	fmt.Printf("L %.5f | R %.5f\n", left, right)
	return nil
}
