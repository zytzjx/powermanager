package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zytzjx/powermanager/crc16"
)

var levelserial *SerialPort

// CRC16Calc modbus
func CRC16Calc(data []byte) (hi byte, low byte, err error) {
	table := crc16.MakeTable(crc16.Crc16MODBUS)
	if table == nil {
		fmt.Printf("Failed to create %q computer\n", crc16.Crc16MODBUS.Name)
		err = fmt.Errorf("failed to create %v computer", crc16.Crc16MODBUS.Name)
		return 0, 0, err
	}
	// 0x534B
	crc := crc16.Checksum(data, table)
	hi = byte(crc & 0xff)
	low = byte((crc & 0xff00) >> 8)
	err = nil
	return
}

func sendSerialbytesData(data []byte, nTimeOut int32) error {
	FDLogger.Printf("\nData:%s, timeout:%d\n", hex.EncodeToString(data), nTimeOut)

	if _, err := levelserial.WriteData(data); err != nil {
		return err
	}
	time.Sleep(10 * time.Microsecond)
	resp, err := levelserial.ReadBytes(nTimeOut)
	if err != nil {
		return err
	}

	if resp != nil && bytes.Equal(data, resp) {
		return nil
	}

	return fmt.Errorf("not found: %s", hex.EncodeToString(resp))
}

func readSerialData(pserial *SerialPort, data []byte, nTimeOut int32, cmd byte) error {
	FDLogger.Printf("\nData:%s, timeout:%d\n", hex.EncodeToString(data), nTimeOut)

	if _, err := pserial.WriteData(data); err != nil {
		return err
	}
	time.Sleep(10 * time.Microsecond)
	resp, err := pserial.ReadBytes(nTimeOut)
	if err != nil {
		return err
	}

	if len(resp) > 4 && resp[1] == cmd {
		return nil
	}

	return fmt.Errorf("not found: %s", hex.EncodeToString(resp))
}

func IsVoltageController(sname string, baudrate int) bool {
	pserial := &SerialPort{mux: &sync.Mutex{}}
	if pserial.Open(sname, baudrate) != nil {
		FDLogger.Printf("Open check port fail: %s, %d\n", sname, baudrate)
		return false
	}
	defer pserial.Close()
	time.Sleep(100 * time.Microsecond)
	return getDeviceModelID(pserial) == nil
}

func getDeviceModelID(pserial *SerialPort) error {
	var ReadPowerSWComand = make([]byte, 8)
	// read Power status
	// 01-03-00-01-00-01-D5-CA
	// ReadPowerSWComand[0] = 0x01
	// ReadPowerSWComand[1] = 0x03
	// ReadPowerSWComand[2] = 0x00
	// ReadPowerSWComand[3] = 0x01
	// ReadPowerSWComand[4] = 0x00
	// ReadPowerSWComand[5] = 0x01
	// ReadPowerSWComand[6] = 0xD5
	// ReadPowerSWComand[7] = 0xCA

	// read model ID
	// 01-03-00-03-00-01-74-0A
	ReadPowerSWComand[0] = 0x01
	ReadPowerSWComand[1] = 0x03
	ReadPowerSWComand[2] = 0x00
	ReadPowerSWComand[3] = 0x03
	ReadPowerSWComand[4] = 0x00
	ReadPowerSWComand[5] = 0x01
	ReadPowerSWComand[6] = 0x74
	ReadPowerSWComand[7] = 0x0A

	if err := readSerialData(pserial, ReadPowerSWComand, 1, ReadPowerSWComand[1]); err != nil {
		FDLogger.Println("set power on Failed")
		return err
	}
	return nil
}

func sendPowerOn() error {
	// 01-06-00-01-00-01-19-CA
	var data = make([]byte, 8)
	data[0] = 0x01
	data[1] = 0x06
	data[2] = 0x00
	data[3] = 0x01
	data[4] = 0x00
	data[5] = 0x01
	data[6] = 0x19
	data[7] = 0xCA
	if err := sendSerialbytesData(data, 1); err != nil {
		FDLogger.Println("set power on Failed")
		return err
	}
	return nil
}

func sendPowerOff() error {
	// 01-06-00-01-00-00-D8-0A
	var data = make([]byte, 8)
	data[0] = 0x01
	data[1] = 0x06
	data[2] = 0x00
	data[3] = 0x01
	data[4] = 0x00
	data[5] = 0x00
	data[6] = 0xD8
	data[7] = 0x0A
	if err := sendSerialbytesData(data, 1); err != nil {
		FDLogger.Println("set power off Failed")
		return err
	}
	return nil
}

func poweroff(c *gin.Context) {
	if err := sendPowerOff(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "power off command failed",
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "OK",
	})
}

func poweron(c *gin.Context) {
	if err := sendPowerOn(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "power on command failed",
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "OK",
	})
}

func voltage(c *gin.Context) {
	v := c.DefaultQuery("v", "1180")
	nv, err := strconv.Atoi(v)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": v,
			"error":   err.Error(),
		})
		return
	}
	if nv > 1250 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": v,
			"error":   "voltage > 12V, LED max voltage is 12V",
		})
		return
	}
	var i int16 = int16(nv)
	// 01-06-00-30-04-C9-4B-53
	var data = make([]byte, 8)
	data[0] = 0x01
	data[1] = 0x06
	data[2] = 0x00
	data[3] = 0x30
	data[4], data[5] = uint8(i>>8), uint8(i&0xff)
	data[6], data[7], _ = CRC16Calc(data[:6])
	err = sendSerialbytesData(data, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "send data failed",
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": hex.EncodeToString(data),
	})
}
