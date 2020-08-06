package main

import (
	"fmt"
	"testing"
)

func TestGetDevUsbList(t *testing.T) {
	usbserials := USBSERIALPORTS{}
	usbserials.GetDevUsbList()
	fmt.Println(usbserials.ttyUSB)
}

func TestGetDevUsbInfo(t *testing.T) {
	usbserials := USBSERIALPORTS{}
	usbserials.getDevUsbInfo("/dev/ttyUSB0")
}
