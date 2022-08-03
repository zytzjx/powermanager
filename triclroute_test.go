package main

import (
	"sync"
	"testing"
)

func TestRecvStatus(t *testing.T) {
	Init()
	powerserial = &SerialPort{mux: &sync.Mutex{}}
	if err := powerserial.Open("/dev/ttyUSB0", 9600); err != nil {
		FDLogger.Fatalf("open power control fail: %s\n", err)
		return
	}
	defer powerserial.Close()
	recvStatus()
}
