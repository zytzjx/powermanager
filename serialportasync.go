package main

import (
	"bytes"
	"encoding/hex"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	// "github.com/jacobsa/go-serial/serial"
	"github.com/tarm/serial"
)

// SerialPort Serial port
type ASerialPort struct {
	// serialopen io.ReadWriteCloser
	serialopen *serial.Port
	mux        *sync.Mutex
	portname   string
	baudrate   int
	IsOpened   bool
	queue      []byte
	muxqueue   *sync.Mutex
}

func (sp *ASerialPort) ClearBuffer() {
	sp.muxqueue.Lock()
	defer sp.muxqueue.Unlock()
	sp.serialopen.Flush()
	sp.queue = sp.queue[0:0]
}

func (sp *ASerialPort) ReadLine() (string, error) {
	sp.muxqueue.Lock()
	defer sp.muxqueue.Unlock()
	if len(sp.queue) == 0 {
		return "", errors.New("buffer is empty")
	}
	if i := bytes.IndexAny(sp.queue, "\r\n"); i >= 0 {
		s := string(sp.queue[0:i])
		sp.queue = sp.queue[i+1:]
		return s, nil
	}
	return "", errors.New("is not a line")
}

func (sp *ASerialPort) ReadLineTimeOut(nTimeout int) (string, error) {
	for start := time.Now(); time.Since(start) < time.Duration(nTimeout)*time.Millisecond; {
		ss, _ := sp.ReadLine()
		if ss != "" {
			FDLogger.Println(ss)
			return ss, nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return "", errors.New("read line timeout")
}

func (sp *ASerialPort) ReadATCmd(cmd string, nTimeOut int) (string, error) {
	FDLogger.Printf("cmd: %s\n", cmd)
	var ss string
	var bcmd bool
	var endcmd bool
	resp := make(chan string)
	var exit bool
	go func() {
		for !exit {
			s, err := sp.ReadLineTimeOut(nTimeOut * 500)
			if err != nil {
				continue
			}
			s1 := strings.TrimSpace(s)
			ss += s1
			ss += "\r\n"
			if strings.Contains(s1, cmd) {
				bcmd = true
			}
			if !bcmd {
				FDLogger.Printf("cmd: %s,(%d)=!=read: %s,(%d)\n", cmd, len(cmd), s1, len(s1))
			}
			if strings.Contains(s1, "OK") || strings.Contains(s1, "ERROR") {
				endcmd = true
				resp <- ss
				break
			}
		}
	}()
	select {
	case strResp := <-resp:
		FDLogger.Println(strResp)
		if !bcmd {
			bcmd = strings.Contains(strResp, cmd)
		}
		if !bcmd && endcmd {
			return strResp, errors.New("not found AT echo")
		}
		return strResp, nil
	case <-time.After(time.Duration(nTimeOut) * time.Second):
		exit = true
		return "", errors.New("recv data timeout")
	}
}

func (sp *ASerialPort) readDataRoutine() {
	buf := make([]byte, 4096)
	for {
		time.Sleep(10 * time.Microsecond)
		n, err := sp.serialopen.Read(buf)
		if err != nil {
			FDLogger.Println("Error reading from serial port: ", err)
			return
		}
		FDLogger.Println(hex.Dump(buf[:n]))
		sp.muxqueue.Lock()
		sp.queue = append(sp.queue, buf[:n]...)
		sp.muxqueue.Unlock()
	}
}

// Open open serial port
func (sp *ASerialPort) Open(PortName string, BaudRate int) error {
	sp.mux = &sync.Mutex{}
	sp.muxqueue = &sync.Mutex{}
	sp.IsOpened = false
	options := &serial.Config{Name: PortName, Baud: BaudRate}

	sp.portname = PortName
	sp.baudrate = BaudRate

	// Open the port.
	// port, err := serial.Open(options)
	port, err := serial.OpenPort(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
		return err
	}
	go sp.readDataRoutine()
	sp.serialopen = port
	sp.IsOpened = true
	return nil
}

// Close port
func (sp *ASerialPort) Close() {
	if sp.IsOpened {
		sp.serialopen.Close()
	}
}

// Flush buffer
func (sp *ASerialPort) Flush() {
	sp.mux.Lock()
	defer sp.mux.Unlock()
	sp.serialopen.Flush()
}

// WriteData port
func (sp *ASerialPort) WriteData(data []byte) (int, error) {
	sp.mux.Lock()
	defer sp.mux.Unlock()
	// FDLogger.Printf("%v\n", data)
	s := string(data)
	FDLogger.Println(s)
	return sp.serialopen.Write(data)
}
