package main

import (
	"bufio"
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

func TestStringLines(t *testing.T) {
	s := "OK"
	if s == "OK" {
		fmt.Println("OK found")
	}
	buf := []byte{65, 84, 71, 48, 13, 10, 65, 84, 87, 48, 13, 10, 65, 84, 87, 48, 13, 10, 65, 84, 71, 48, 13, 10, 82, 85, 78, 78, 73, 78, 71, 13, 10, 65, 84, 87, 49, 13, 10, 65, 84, 71, 45, 49, 13, 79, 75, 13, 10}
	bytesReader := bytes.NewReader(buf)
	line := bufio.NewScanner(bytesReader)
	line.Split(ScanItems)
	for line.Scan() {
		fmt.Println(line.Text())
	}
	/*
		ATG0
		ATW0
		ATW0
		ATG0
		RUNNING
		ATW1
		ATG-1
		OK
		PASS
	*/
}

func TestByteStringCheck(t *testing.T) {
	buf := []byte{65, 84, 71, 48, 13, 10, 65, 84, 87, 48, 13, 10, 65, 84, 87, 48, 13, 10, 65, 84, 71, 48, 13, 10, 82, 85, 78, 78, 73, 78, 71, 13, 10, 65, 84, 87, 49, 13, 10, 65, 84, 71, 45, 49, 13, 79, 75, 13, 10}
	bytesReader := bytes.NewReader(buf)
	bufReader := bufio.NewReader(bytesReader)
	for {
		value1, _, err := bufReader.ReadLine()
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Println(string(value1))
	}
}
func TestByteArray(t *testing.T) {
	buf := []byte{65, 84, 71, 48, 13, 10, 65, 84, 87, 48, 13, 10, 65, 84, 87, 48, 13, 10, 65, 84, 71, 48, 13, 10, 82, 85, 78, 78, 73, 78, 71, 13, 10, 65, 84, 87, 49, 13, 10, 65, 84, 71, 45, 49, 13, 79, 75, 13, 10}
	if bytes.Contains(buf, []byte("OK\r")) || bytes.Contains(buf, []byte("ERROR")) {
		t.Log("found")
		return
	}
	t.Error("not found")
}

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

func TestDevsInfo(t *testing.T) {
	Init()
	infos, err := DevsInfo("/dev/ttyUSB0")
	if err != nil {
		t.Error(err)
	}
	AssertNotEqual(t, len(infos), 0)
}
