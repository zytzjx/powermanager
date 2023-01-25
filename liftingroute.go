package main

import (
	"bufio"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var liftingserial *ASerialPort
var liftMutex = &sync.Mutex{}
var muxWait = &sync.Mutex{}

// AT COMMAND LOCK TIME
var INTERVAL = 300

// TurnFlik record flik
type TurnFlik struct {
	cmd   string
	bPlus bool
	mux   *sync.Mutex
}

// NewTurnFlik create this struct
func NewTurnFlik(command string) *TurnFlik {
	tf := TurnFlik{cmd: command, bPlus: true, mux: &sync.Mutex{}}
	return &tf
}

// get cmd
func (tf *TurnFlik) getCmd() string {
	tf.mux.Lock()
	defer tf.mux.Unlock()
	if tf.bPlus {
		tf.bPlus = !tf.bPlus
		return fmt.Sprintf("%s+\r", tf.cmd)
	}
	tf.bPlus = !tf.bPlus
	return fmt.Sprintf("%s-\r", tf.cmd)

}

// get cmd
func (tf *TurnFlik) setPlus(b bool) {
	tf.mux.Lock()
	defer tf.mux.Unlock()
	tf.bPlus = b
}

var flipcmd *TurnFlik = NewTurnFlik("ATF")
var turncmd *TurnFlik = NewTurnFlik("ATT")

func sendSerialData(cmd string, nTimeOut int32, interval int, delay int) (string, error) {
	FDLogger.Printf("\ncmd:%s, timeout:%d, interval:%d\n", cmd, nTimeOut, interval)
	liftingserial.ClearBuffer()
	bbcmd := []byte(cmd)
	if _, err := liftingserial.WriteData(bbcmd); err != nil {
		return "", err
	}
	time.Sleep(time.Duration(delay) * time.Millisecond)
	resp, err := liftingserial.ReadATCmd(strings.TrimSpace(cmd), int(nTimeOut))
	if err != nil {
		return "", err
	}
	FDLogger.Printf("cmd:%s--> resp: %s\n", cmd, resp)
	if strings.Contains(resp, "OK") {
		return resp, nil
	}
	return "", fmt.Errorf("not found OK: %s", resp)
}

/*
func sendSerialData_(cmd string, nTimeOut int32, interval int, delay int) (string, error) {
	FDLogger.Printf("\ncmd:%s, timeout:%d, interval:%d\n", cmd, nTimeOut, interval)
	liftingserial.serialopen.Flush()
	bbcmd := []byte(cmd)
	for i := 0; i < len(bbcmd); i++ {
		if _, err := liftingserial.WriteData(bbcmd[i : i+1]); err != nil {
			return "", err
		}
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
	time.Sleep(time.Duration(delay) * time.Millisecond)
	resp, err := liftingserial.ReadData(cmd, nTimeOut)
	if err != nil {
		return "", err
	}
	FDLogger.Printf("cmd:%s--> resp: %s\n", cmd, resp)
	if strings.Contains(resp, "OK") {
		return resp, nil
	}
	return "", fmt.Errorf("not found OK: %s", resp)
}
*/
// func sendSerialDataATC(cmd string, nTimeOut int32, interval int, delay int) (string, error) {
// 	FDLogger.Printf("\ncmd:%s, timeout:%d, interval:%d\n", cmd, nTimeOut, interval)
// 	bbcmd := []byte(cmd)
// 	for i := 0; i < len(bbcmd); i++ {
// 		if _, err := liftingserial.WriteData(bbcmd[i : i+1]); err != nil {
// 			return "", err
// 		}
// 		time.Sleep(time.Duration(interval) * time.Millisecond)
// 	}
// 	time.Sleep(time.Duration(delay) * time.Millisecond)
// 	resp, err := liftingserial.ReadDataATC(nTimeOut)
// 	if err != nil {
// 		return "", err
// 	}
// 	FDLogger.Printf("cmd:%s--> resp: %s\n", cmd, resp)
// 	if strings.Contains(resp, "OK") {
// 		return resp, nil
// 	}
// 	return "", fmt.Errorf("not found OK: %s", resp)
// }

func setNeedSleepFlag() {
	muxWait.Lock()
	// muxWait.Unlock()
}

func waitSleepFlag() {
	// muxWait.Lock()
	time.Sleep(time.Duration(INTERVAL) * time.Millisecond)
	muxWait.Unlock()
}

func hello(c *gin.Context) {
	liftMutex.Lock()
	defer liftMutex.Unlock()

	defer func() {
		go waitSleepFlag()
	}()
	setNeedSleepFlag()

	cmd := "AT\r"
	SendATCmd(cmd, c)
}

func status(c *gin.Context) {
	liftMutex.Lock()
	defer liftMutex.Unlock()

	defer func() {
		go waitSleepFlag()
	}()
	setNeedSleepFlag()

	cmd := "ATS\r"
	SendATCmd(cmd, c)
}

func information(c *gin.Context) {
	liftMutex.Lock()
	defer liftMutex.Unlock()

	defer func() {
		go waitSleepFlag()
	}()
	setNeedSleepFlag()

	cmd := "ATI\r"
	sinterval := c.DefaultQuery("interval", "1")
	interval, err := strconv.Atoi(sinterval)
	if err != nil {
		interval = 1
	}
	sdelay := c.DefaultQuery("delay", "10")
	ndelay, err := strconv.Atoi(sdelay)
	if err != nil {
		ndelay = 10
	}
	resp, err := sendSerialData(cmd, 6, interval, ndelay)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	scanner := bufio.NewScanner(strings.NewReader(resp))
	parse := make(map[string]string)
	for scanner.Scan() {
		x := scanner.Text()
		keyvalue := strings.Split(x, "=")
		if len(keyvalue) == 2 && keyvalue[0] != "" {
			parse[keyvalue[0]] = keyvalue[1]
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": resp,
		"parse":   parse,
	})
}

func stop(c *gin.Context) {
	liftMutex.Lock()
	defer liftMutex.Unlock()

	defer func() {
		go waitSleepFlag()
	}()
	setNeedSleepFlag()

	cmd := "ATSTOP\r"
	SendATCmd(cmd, c)
}

func reset(c *gin.Context) {
	liftMutex.Lock()
	defer liftMutex.Unlock()

	defer func() {
		go waitSleepFlag()
	}()
	setNeedSleepFlag()

	cmd := "ATZ\r"
	SendATCmd(cmd, c)
}

func getlasterror(c *gin.Context) {
	liftMutex.Lock()
	defer liftMutex.Unlock()

	defer func() {
		go waitSleepFlag()
	}()
	setNeedSleepFlag()

	cmd := "ATE?\r"
	SendATCmd(cmd, c)
}

func start(c *gin.Context) {
	liftMutex.Lock()
	defer liftMutex.Unlock()

	defer func() {
		go waitSleepFlag()
	}()
	setNeedSleepFlag()

	cmd := "ATSTART\r"
	SendATCmd(cmd, c)
}

func home(c *gin.Context) {
	liftMutex.Lock()
	defer liftMutex.Unlock()

	defer func() {
		go waitSleepFlag()
	}()
	setNeedSleepFlag()

	cmd := "ATC%s\r"
	flag := c.DefaultQuery("flag", "2")
	cmd = fmt.Sprintf(cmd, flag)
	SendATCmd(cmd, c)
}

func wind(c *gin.Context) {
	liftMutex.Lock()
	defer liftMutex.Unlock()

	defer func() {
		go waitSleepFlag()
	}()
	setNeedSleepFlag()

	cmd := "ATW%s\r"
	flag := c.DefaultQuery("flag", "0")
	cmd = fmt.Sprintf(cmd, flag)
	SendATCmd(cmd, c)
}

func carrier(c *gin.Context) {
	liftMutex.Lock()
	defer liftMutex.Unlock()

	defer func() {
		go waitSleepFlag()
	}()
	setNeedSleepFlag()

	cmd := "ATX%s\r"
	flag := c.DefaultQuery("flag", "0")
	cmd = fmt.Sprintf(cmd, flag)
	SendATCmd(cmd, c)
}

// PositionInfo position info
type PositionInfo struct {
	Position int    `form:"p"`
	Value    string `form:"value"`
}

func goposition(c *gin.Context) {
	liftMutex.Lock()
	defer liftMutex.Unlock()

	defer func() {
		go waitSleepFlag()
	}()
	setNeedSleepFlag()

	cmd := "ATG%d\r"
	var pp PositionInfo

	if c.ShouldBindQuery(&pp) == nil {
		cmd = fmt.Sprintf("ATG%d\r", pp.Position)
		SendATCmd(cmd, c)
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{
		"status":  "Error",
		"message": "go posion number",
	})
}

func setPoisition(c *gin.Context) {
	liftMutex.Lock()
	defer liftMutex.Unlock()

	defer func() {
		go waitSleepFlag()
	}()
	setNeedSleepFlag()

	var pp PositionInfo
	sinterval := c.DefaultQuery("interval", "1")
	interval, err := strconv.Atoi(sinterval)
	if err != nil {
		interval = 1
	}
	sdelay := c.DefaultQuery("delay", "10")
	ndelay, err := strconv.Atoi(sdelay)
	if err != nil {
		ndelay = 10
	}

	if c.ShouldBindQuery(&pp) == nil {
		val, err := strconv.ParseFloat(pp.Value, 64)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"data":  fmt.Sprintf("%s is not float string", pp.Value),
			})
			return
		}
		var value string
		if val < 0 {
			value = fmt.Sprintf("%07.2f", val)
		} else {
			value = fmt.Sprintf("+%06.2f", val)
		}
		cmd := fmt.Sprintf("ATP%d=%s\r", pp.Position, value)
		FDLogger.Println(cmd)
		FDLogger.Printf("len=%d\n", len(cmd))
		// cmd = "ATP6=+370.00\r"
		resp, err := sendSerialData(cmd, 3, interval, ndelay)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  "OK",
			"message": resp,
		})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{
		"status":  "Error",
		"message": "set posion number & value",
	})
}

func queryAirValue(c *gin.Context) {
	liftMutex.Lock()
	defer liftMutex.Unlock()

	defer func() {
		go waitSleepFlag()
	}()
	setNeedSleepFlag()
	cmd := "ATQ\r"
	sinterval := c.DefaultQuery("interval", "1")
	interval, err := strconv.Atoi(sinterval)
	if err != nil {
		interval = 1
	}
	sdelay := c.DefaultQuery("delay", "10")
	ndelay, err := strconv.Atoi(sdelay)
	if err != nil {
		ndelay = 10
	}
	resp, err := sendSerialData(cmd, 3, interval, ndelay)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	value := strings.Replace(resp, "ATQ", "", -1)
	value = strings.Replace(value, "OK", "", -1)
	value = strings.Replace(value, "\r\n", "", -1)
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": resp,
		"value":   value,
	})
}

func listposition(c *gin.Context) {
	liftMutex.Lock()
	defer liftMutex.Unlock()

	defer func() {
		go waitSleepFlag()
	}()
	setNeedSleepFlag()

	cmd := "ATP?\r"
	var pp PositionInfo
	if c.Query("p") != "" {
		if c.ShouldBindQuery(&pp) == nil {
			cmd = fmt.Sprintf("ATP%d?\r", pp.Position)
		}
	}
	sinterval := c.DefaultQuery("interval", "1")
	interval, err := strconv.Atoi(sinterval)
	if err != nil {
		interval = 1
	}
	sdelay := c.DefaultQuery("delay", "10")
	ndelay, err := strconv.Atoi(sdelay)
	if err != nil {
		ndelay = 10
	}
	resp, err := sendSerialData(cmd, 3, interval, ndelay)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(resp))
	parse := make(map[string]string)
	for scanner.Scan() {
		x := scanner.Text()
		keyvalue := strings.Split(x, "=")
		if len(keyvalue) == 2 && keyvalue[0] != "" {
			parse[keyvalue[0]] = keyvalue[1]
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": resp,
		"parse":   parse,
	})
}

func turn(c *gin.Context) {
	liftMutex.Lock()
	defer liftMutex.Unlock()

	defer func() {
		go waitSleepFlag()
	}()
	setNeedSleepFlag()

	cmd := "ATT%s\r"
	value := c.DefaultQuery("flag", "+")
	switch value {
	case "+":
		cmd = fmt.Sprintf(cmd, value)
		turncmd.setPlus(false)
	case "-":
		cmd = fmt.Sprintf(cmd, value)
		turncmd.setPlus(true)
	default:
		cmd = turncmd.getCmd()
	}
	SendATCmd(cmd, c)
}

func flip(c *gin.Context) {
	liftMutex.Lock()
	defer liftMutex.Unlock()

	defer func() {
		go waitSleepFlag()
	}()
	setNeedSleepFlag()

	cmd := "ATF%s\r"
	value := c.DefaultQuery("flag", "+")
	switch value {
	case "+":
		cmd = fmt.Sprintf(cmd, value)
		flipcmd.setPlus(false)
	case "-":
		cmd = fmt.Sprintf(cmd, value)
		flipcmd.setPlus(true)
	default:
		cmd = flipcmd.getCmd()
	}
	SendATCmd(cmd, c)
}

func SendATCmd(cmd string, c *gin.Context) {
	sinterval := c.DefaultQuery("interval", "1")
	interval, err := strconv.Atoi(sinterval)
	if err != nil {
		interval = 1
	}
	sdelay := c.DefaultQuery("delay", "10")
	ndelay, err := strconv.Atoi(sdelay)
	if err != nil {
		ndelay = 10
		if strings.HasPrefix(cmd, "ATC") {
			ndelay = 15000
		}
	}
	resp, err := sendSerialData(cmd, 10, interval, ndelay)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": resp,
	})
}
