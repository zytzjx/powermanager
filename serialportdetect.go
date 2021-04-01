package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
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
	// "github.com/tarm/serial"
)

// SerialPort Serial port
type SerialPort struct {
	serialopen io.ReadWriteCloser
	// serialopen *serial.Port
	mux      *sync.Mutex
	portname string
	baudrate int
	IsOpened bool
}

var nRecordATCError int

func ScanItems(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexAny(data, "\r\n"); i >= 0 {
		return i + 1, data[0:i], nil
	}

	if atEOF {
		return len(data), data, nil
	}

	return 0, nil, nil
}

// Open open serial port
func (sp *SerialPort) Open(PortName string, BaudRate int) error {
	sp.IsOpened = false
	// options := &serial.Config{Name: PortName, Baud: BaudRate}
	options := serial.OpenOptions{
		PortName:        PortName,
		BaudRate:        uint(BaudRate),
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}
	sp.portname = PortName
	sp.baudrate = BaudRate

	// Open the port.
	port, err := serial.Open(options)
	// port, err := serial.OpenPort(options)
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
	// FDLogger.Printf("%v\n", data)
	s := string(data)
	FDLogger.Println(s)
	return sp.serialopen.Write(data)
}

// ReadBytes read bytes from serial port
func (sp *SerialPort) ReadBytes(nTimeout int32) ([]byte, error) {
	resp := make(chan []byte)
	err := make(chan error)
	go func(resp chan []byte, errr chan error) {
		buf := make([]byte, 4096)
		cnt := 0
		for {
			time.Sleep(10 * time.Microsecond)
			n, err := func() (int, error) {
				return sp.serialopen.Read(buf[cnt:])
			}()

			if err != nil {
				FDLogger.Println("Error reading from serial port: ", err)
				errr <- err
				return
			}
			cnt += n
			FDLogger.Println(buf[0:cnt])
			if cnt >= 6 { // this is modbus: id cmd addr(2) crc(2) , so >=6.
				break
			}
		}

		resp <- buf[:cnt]

	}(resp, err)

	select {
	case strResp := <-resp:
		FDLogger.Println(strResp)
		return strResp, nil
	case errret := <-err:
		FDLogger.Println(errret)
		return nil, errret
	case <-time.After(time.Duration(nTimeout) * time.Second):
		return nil, errors.New("recv data timeout")
	}

}

// ReadDataLen read from usb port, timeout is microsecond
func (sp *SerialPort) ReadDataLen(nTimeout int32) (string, error) {
	resp := make(chan string)
	errchan := make(chan error)
	go func() {
		buf := make([]byte, 4096)
		cnt := 0
		for {
			time.Sleep(10 * time.Microsecond)
			n, err := func() (int, error) {
				return sp.serialopen.Read(buf[cnt:])
			}()

			if err != nil {
				FDLogger.Println("Error reading from serial port: ", err)
				errchan <- err
				return
			}
			cnt += n
			FDLogger.Println(hex.Dump(buf[0:cnt]))
			if bytes.Contains(buf, []byte("\r\n")) {
				break
			}
		}

		resp <- string(buf[:cnt])

	}()

	select {
	case strResp := <-resp:
		FDLogger.Println(strResp)
		return strResp, nil
	case errret := <-errchan:
		FDLogger.Println(errret)
		return "", errret
	case <-time.After(time.Duration(nTimeout) * time.Second):
		return "", errors.New("recv data timeout")
	}
}

// ReadData read from usb port
func (sp *SerialPort) ReadDataEnd(nTimeout int32) (string, error) {
	resp := make(chan string)
	err := make(chan error)
	go func(resp chan string, errr chan error) {
		buf := make([]byte, 4096)
		cnt := 0
		for {
			time.Sleep(10 * time.Microsecond)
			n, err := func() (int, error) {
				// sp.mux.Lock()
				// defer sp.mux.Unlock()
				return sp.serialopen.Read(buf[cnt:])
			}()

			if err != nil {
				FDLogger.Println("Error reading from serial port: ", err)
				errr <- err
				return
			}
			cnt += n
			FDLogger.Println(hex.Dump(buf[0:cnt]))
			bytesReader := bytes.NewReader(buf[0:cnt])
			line := bufio.NewScanner(bytesReader)
			line.Split(ScanItems)
			var found bool
			for line.Scan() {
				s := line.Text()
				FDLogger.Println(s)
				if s == "OK" || s == "ERROR" {
					found = true
					break
				}
			}

			if found || bytes.Contains(buf, []byte("OK")) || bytes.Contains(buf, []byte("ERROR")) {
				break
			}
		}

		resp <- string(buf[:cnt])

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

func (sp *SerialPort) Read(buf []byte) (n int, err error) {
	ch := make(chan bool)
	n = 0
	err = nil
	go func() {
		n, err = sp.serialopen.Read(buf)
		ch <- true
	}()
	select {
	case <-ch:
		return
	case <-time.After(5 * time.Second):
		return 0, errors.New("Timeout")
	}
}

// ReadData read from usb port
func (sp *SerialPort) ReadData(cmd string, nTimeout int32) (string, error) {
	resp := make(chan string)
	err := make(chan error)
	cmd1 := strings.TrimSpace(cmd)
	bExitTimout := false
	go func(resp chan string, errr chan error) {
		buf := make([]byte, 4096)
		cnt := 0
		for {
			time.Sleep(10 * time.Microsecond)
			// n, err := func() (int, error) {
			// 	return sp.serialopen.Read(buf[cnt:])
			// }()
			n, err := sp.Read(buf[cnt:])
			if err != nil && bExitTimout {
				FDLogger.Println("Error reading from serial port: ", err)
				errr <- err
				return
			}
			cnt += n

			FDLogger.Println(hex.Dump(buf[0:cnt]))

			bytesReader := bytes.NewReader(buf[0:cnt])
			line := bufio.NewScanner(bytesReader)
			line.Split(ScanItems)
			var found bool
			var foundcmd bool
			for line.Scan() {
				s := strings.TrimSpace(line.Text())
				FDLogger.Println(cmd1 + "    " + s)
				if s == "OK" || s == "ERROR" {
					found = true
				} else if s == cmd1 {
					foundcmd = true
				}
			}

			if found && foundcmd { //|| bytes.Contains(buf, []byte("OK")) || bytes.Contains(buf, []byte("ERROR"))
				FDLogger.Println("find correct protocol")
				break
			}
			if found && !foundcmd {
				FDLogger.Println("find end but not found command")
				errr <- errors.New("communication protocol failed")
				return
			}
		}
		resp <- string(buf[:cnt])
	}(resp, err)

	select {
	case strResp := <-resp:
		FDLogger.Println(strResp)
		return strResp, nil
	case errret := <-err:
		FDLogger.Println(errret)
		return "", errret
	case <-time.After(time.Duration(nTimeout) * time.Second):
		bExitTimout = true
		return "", errors.New("recv data timeout")
	}
}

// USBSERIALPORTS serial ports configs
type USBSERIALPORTS struct {
	Power      string `json:"power"`
	Lifting    string `json:"lifting"`
	Level      string `json:"voltage"`
	PBaudRate  int    `json:"powerbaudrate"`
	LBaudRate  int    `json:"liftingbaudrate"`
	LevelBRate int    `json:"voltagebaudrate"`

	serialPower   string
	serialLifting string
	serialVoltage string
	ttyUSB        []string
}

func retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; ; i++ {
		err = f()
		if err == nil {
			return
		}

		if i >= (attempts - 1) {
			break
		}

		time.Sleep(sleep)

		log.Println("retrying after error:", err)
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}

// IsPowerSerial check is power serial
func IsPowerSerial(sname string, baudrate int) bool {
	pserial := &SerialPort{mux: &sync.Mutex{}}
	if pserial.Open(sname, baudrate) != nil {
		FDLogger.Printf("Open check port fail: %s, %d\n", sname, baudrate)
		return false
	}
	defer pserial.Close()
	time.Sleep(100 * time.Microsecond)
	var ret string
	errretry := retry(3, 10*time.Microsecond, func() error {
		_, err := pserial.WriteData([]byte("?\r"))
		if err != nil {
			FDLogger.Printf("IsPowerSerial Failed to write data: %s\n", err)
			return err
		}
		ret, err = pserial.ReadDataLen(1)
		if err != nil {
			FDLogger.Printf("IsPowerSerial Failed to readdata: %s\n", err)
			return err
		}
		return nil
	})

	if errretry != nil {
		return false
	}

	if strings.HasPrefix(ret, "I,") || strings.HasPrefix(ret, "POWER") {
		return true
	}
	FDLogger.Printf("IsPowerSerial readdata: %s\n", ret)
	return false
}

/*
func (usp *USBSERIALPORTS) getDevBus(path string) (int, int, error) {
	pathes := strings.Split(path, "-")
	if len(pathes) < 2 {
		return 0, 0, errors.New("usb device is not BUS. one is BUS")
	}
	var nDev int
	busindex, err := strconv.Atoi(pathes[0])
	if err != nil {
		return 0, 0, err
	}
	nDev, err = usbview.GetDevFromPathes(pathes)
	if err != nil {
		return 0, 0, err
	}
	return busindex, nDev, nil
}
*/
// LoadConfig load config from config file
func (usp *USBSERIALPORTS) LoadConfig(filename string) error {
	// RunLsusb()
	if filename == "" {
		filename = "serialcalibration.json"
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

// DevsInfo get Device name information
func DevsInfo(DevName string) (map[string]string, error) {
	infos := map[string]string{}
	cmd := fmt.Sprintf("udevadm info -n %s", DevName)
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		FDLogger.Printf("Failed to execute command: %s, %s\n", cmd, err)
		return infos, err
	}
	// FDLogger.Printf("result execute command: %s, %s\n", cmd, string(out))
	scanner := bufio.NewScanner(bytes.NewReader(out))
	re := regexp.MustCompile(`^.: (.*?)=(.*)$`)
	for scanner.Scan() {
		x := scanner.Text()
		substr := re.FindStringSubmatch(x)
		if len(substr) == 3 {
			infos[substr[1]] = substr[2]
		}
	}
	return infos, nil
}

// LoadUSBDevsWithoutConfig if not found USB config
func (usp *USBSERIALPORTS) LoadUSBDevsWithoutConfig() error {
	if err := usp.GetDevUsbList(); err != nil {
		return err
	}
	if len(usp.ttyUSB) == 0 {
		return errors.New("not find ttyUSB serial port")
	}
	// 1 or 2 USB
	if len(usp.ttyUSB) < 4 {
		for _, devname := range usp.ttyUSB {
			infos, err := DevsInfo(devname)
			if err != nil {
				return err
			}
			var bch340Vid, bch340Pid bool
			var bVid, bPid bool
			if vid, ok := infos["ID_VENDOR_ID"]; ok {
				if vid == "1a86" {
					bch340Vid = true
				} else if vid == "0403" {
					bVid = true
				}
			}
			if pid, ok := infos["ID_MODEL_ID"]; ok {
				if pid == "7523" {
					bch340Pid = true
				} else if pid == "6001" {
					bPid = true
				}
			}
			if bch340Vid && bch340Pid {
				if usp.serialVoltage == "" && IsVoltageController(devname, 9600) {
					usp.serialVoltage = devname
					usp.LevelBRate = 9600
					FDLogger.Printf("Found power supply: %s\n", devname)
				} else {
					usp.serialPower = devname
					usp.PBaudRate = 9600
					FDLogger.Printf("Found arduino: %s\n", devname)
				}
				//  if usp.serialPower == "" && IsPowerSerial(devname, 9600) {
				// 	usp.serialPower = devname
				// 	usp.PBaudRate = 9600
				// 	FDLogger.Printf("Found arduino: %s\n", devname)
				// } else {
				// 	usp.serialVoltage = devname
				// 	usp.LevelBRate = 9600
				// 	FDLogger.Printf("Found power supply: %s\n", devname)
				// }

			}
			if bVid && bPid {
				usp.serialLifting = devname
				usp.LBaudRate = 115200
				FDLogger.Printf("Found lifting: %s\n", devname)
			}
		}
	} else {
		return errors.New("too much USB Serial ports are found")
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

// getDeviceLocationpath get device path
func (usp *USBSERIALPORTS) getDeviceLocationpath(devName string) (string, error) {
	cmd := `udevadm info -q path -n %s`
	cmd = fmt.Sprintf(cmd, devName)
	fmt.Println(cmd)
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		FDLogger.Printf("Failed to execute command: %s\n", err)
		return "", err
	}
	re := regexp.MustCompile(`.*/(.*?)/(.*?):.*?/ttyUSB\\d+/`)
	pathes := re.FindStringSubmatch(string(out))
	if len(pathes) > 2 {
		return pathes[1], nil
	}
	return "", fmt.Errorf("path return is: %s", string(out))
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

// VerifyDevName from location path
func (usp *USBSERIALPORTS) VerifyDevName() error {
	if err := usp.GetDevUsbList(); err != nil {
		return err
	}
	for _, s := range usp.ttyUSB {
		locpath, err := usp.getDeviceLocationpath(s)
		if err != nil {
			FDLogger.Println(err)
			continue
		}
		locpath = strings.Replace(locpath, ".", "-", -1)
		if locpath == usp.Power {
			usp.serialPower = s
			FDLogger.Printf("found power serial: %s\n", s)
		} else if locpath == usp.Lifting {
			usp.serialLifting = s
			FDLogger.Printf("found lifting serial: %s\n", s)
		} else if locpath == usp.Level {
			usp.serialVoltage = s
			FDLogger.Printf("found voltage serial: %s\n", s)
		}
	}
	return nil
}

/*
func (usp *USBSERIALPORTS) verifyDevName() error {
	if err := usp.GetDevUsbList(); err != nil {
		return err
	}
	powerbus, powerdev, err := usp.getDevBus(usp.Power)
	if err != nil {
		FDLogger.Printf("power parser error: %s\n", err)
		return err
	}
	voltagebus, voltagedev, err := usp.getDevBus(usp.Level)
	if err != nil {
		FDLogger.Printf("voltage parser error: %s\n", err)
		return err
	}
	liftbus, liftdev, err := usp.getDevBus(usp.Lifting)
	if err != nil {
		FDLogger.Printf("lift parser error: %s\n", err)
		return err
	}

	for _, s := range usp.ttyUSB {
		devnum, busnum, err := usp.getDevUsbInfo(s)
		if err != nil {
			FDLogger.Printf("error: %s\n", err)
			continue
		}
		nbusnum, _ := strconv.Atoi(busnum)
		ndevnum, _ := strconv.Atoi(devnum)
		if nbusnum == powerbus && ndevnum == powerdev {
			usp.serialPower = s
			FDLogger.Printf("found power serial: %s\n", s)
		} else if nbusnum == liftbus && ndevnum == liftdev {
			usp.serialLifting = s
			FDLogger.Printf("found lifting serial: %s\n", s)
		} else if nbusnum == voltagebus && ndevnum == voltagedev {
			usp.serialVoltage = s
			FDLogger.Printf("found power serial: %s\n", s)
		}
		// sstart := fmt.Sprintf("Bus %s Device %s: ID", busnum, devnum)
		// if strings.HasPrefix(usp.Power, sstart) {
		// 	usp.serialPower = s
		// 	FDLogger.Printf("found power serial: %s\n", s)
		// } else if strings.HasPrefix(usp.Lifting, sstart) {
		// 	usp.serialLifting = s
		// 	FDLogger.Printf("found lifting serial: %s\n", s)
		// }
	}
	return nil
}
*/
