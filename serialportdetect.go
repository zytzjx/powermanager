package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

// SerialPort Serial port
type SerialPort struct {
	serialopen io.ReadWriteCloser
	mux        *sync.Mutex
	portname   string
	baudrate   int
	IsOpened   bool
}

// Open open serial port
func (sp *SerialPort) Open(PortName string, BaudRate int) error {
	sp.IsOpened = false
	options := serial.OpenOptions{
		PortName:        PortName,
		BaudRate:        9600,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}
	sp.portname = PortName
	sp.baudrate = BaudRate
	sp.mux = &sync.Mutex{}
	// Open the port.
	port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
		return err
	}
	sp.serialopen = port
	sp.IsOpened = true
	return nil
}

// Close port
func (sp *SerialPort) Close() {
	if sp.IsOpened {
		sp.serialopen.Close()
	}
}

// WriteData port
func (sp *SerialPort) WriteData(data []byte) (int, error) {
	sp.mux.Lock()
	defer sp.mux.Unlock()
	return sp.serialopen.Write(data)
}

// ReadData read from usb port
func (sp *SerialPort) ReadData(nTimeout int32) (string, error) {
	resp := make(chan string)
	err := make(chan error)
	go func(resp chan string, errr chan error) {
		buf := make([]byte, 4096)
		sp.mux.Lock()
		n, err := sp.serialopen.Read(buf)
		sp.mux.Unlock()
		if err != nil {
			FDLogger.Println("Error reading from serial port: ", err)
			errr <- err
			return
		}
		resp <- string(buf[:n])

	}(resp, err)

	select {
	case strResp := <-resp:
		FDLogger.Println(strResp)
		return strResp, nil
	case errret := <-err:
		FDLogger.Println(errret)
		return "", errret
	case <-time.After(time.Duration(nTimeout) * time.Second):
		return "", errors.New("recv data timeout")
	}
}

// USBSERIALPORTS serial ports configs
type USBSERIALPORTS struct {
	Power   string `json:"power"`
	Lifting string `json:"lifting"`

	serialPower   string
	serialLifting string
	ttyUSB        []string
}

// LoadConfig load config from config file
func (usp *USBSERIALPORTS) LoadConfig(filename string) error {
	if filename == "" {
		filename = "calibration.json"
	}
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(file), usp)
	if err != nil {
		return nil
	}
	return nil
}

//GetDevUsbList List ttyUSB* in System
func (usp *USBSERIALPORTS) GetDevUsbList() error {
	// return string(out)
	cmd := "ls -l /dev/ttyUSB*"
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		FDLogger.Printf("Failed to execute command: %s, %s\n", cmd, err)
		return err
	}
	FDLogger.Printf("result execute command: %s, %s\n", cmd, string(out))
	re := regexp.MustCompile(`/dev/ttyUSB\d+`)
	ss := string(out)
	usp.ttyUSB = re.FindAllString(ss, -1)
	return nil
}

func (usp *USBSERIALPORTS) getDevUsbInfo(devName string) (string, string, error) {
	// echo `udevadm info --name=/dev/ttyUSB0 --attribute-walk | sed -n 's/\s*ATTRS{\(\(devnum\)\|\(busnum\)\)}==\"\([^\"]\+\)\"/\1\ \4/p' | head -n 2 | awk '{$1 = sprintf("%s:%03d", $1, $2); print $1;}'`
	cmd := `udevadm info --name=%s --attribute-walk | sed -n 's/\s*ATTRS{\(\(devnum\)\|\(busnum\)\)}==\"\([^\"]\+\)\"/\1\ \4/p' | head -n 2 | awk '{$1 = sprintf("%%s:%%03d", $1, $2); print $1;}'`
	cmd = fmt.Sprintf(cmd, devName)
	fmt.Println(cmd)
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		FDLogger.Printf("Failed to execute command: %s\n", err)
		return "", "", err
	}
	//fmt.Println(string(out))
	var devnum, busnum string
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		x := scanner.Text()
		nums := strings.Split(x, ":")
		if len(nums) == 2 {
			switch nums[0] {
			case "devnum":
				devnum = nums[1]
			case "busnum":
				busnum = nums[1]
			default:
				FDLogger.Println("out put error, the result is not aspect.")
			}
		} else {
			FDLogger.Printf("output data format, the result is not aspect. %s\n", x)
		}

	}

	return devnum, busnum, nil
}

func (usp *USBSERIALPORTS) verifyDevName() error {
	if err := usp.GetDevUsbList(); err != nil {
		return err
	}
	for _, s := range usp.ttyUSB {
		devnum, busnum, err := usp.getDevUsbInfo(s)
		if err != nil {
			FDLogger.Printf("error: %s\n", err)
			continue
		}
		sstart := fmt.Sprintf("Bus %s Device %s: ID", busnum, devnum)
		if strings.HasPrefix(usp.Power, sstart) {
			usp.serialPower = s
			FDLogger.Printf("found power serial: %s\n", s)
		} else if strings.HasPrefix(usp.Lifting, sstart) {
			usp.serialLifting = s
			FDLogger.Printf("found lifting serial: %s\n", s)
		}
	}
	return nil
}
