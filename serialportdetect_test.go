package main

import (
	"fmt"
	"reflect"
	"testing"
)

// AssertEqual checks if values are equal
func AssertEqual(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		return
	}
	// debug.PrintStack()
	t.Errorf("Received %v (type %v), expected %v (type %v)", a, reflect.TypeOf(a), b, reflect.TypeOf(b))
}

// AssertEqual checks if values are equal
func AssertNotEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		return
	}
	// debug.PrintStack()
	t.Errorf("Received %v (type %v), expected %v (type %v)", a, reflect.TypeOf(a), b, reflect.TypeOf(b))
}

func TestGetDevUsbList(t *testing.T) {
	usbserials := USBSERIALPORTS{}
	usbserials.GetDevUsbList()
	fmt.Println(usbserials.ttyUSB)
}

func TestGetDevUsbInfo(t *testing.T) {
	usbserials := USBSERIALPORTS{}
	usbserials.getDevUsbInfo("/dev/ttyUSB0")
}

func TestVerifyDevName(t *testing.T) {
	Init()
	usbserialList := USBSERIALPORTS{}
	usbserialList.LoadConfig("serialcalibration.json")
	if err := usbserialList.VerifyDevName(); err != nil {
		t.Error(err)
		return
	}
	AssertNotEqual(t, usbserialList.serialLifting, "")
	AssertNotEqual(t, usbserialList.serialPower, "")
}
