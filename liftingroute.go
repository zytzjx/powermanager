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

var liftingserial *SerialPort
var liftMutex = &sync.Mutex{}
var muxWait = &sync.Mutex{}
var bNeedSleep = false

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

func sendSerialData(cmd string, nTimeOut int32, interval int) (string, error) {
	FDLogger.Printf("\ncmd:%s, timeout:%d, interval:%d\n", cmd, nTimeOut, interval)
	bbcmd := []byte(cmd)
	for i := 0; i < len(bbcmd); i++ {
		if _, err := liftingserial.WriteData(bbcmd[i : i+1]); err != nil {
			return "", err
		}
		time.Sleep(time.Duration(interval) * time.Microsecond)
	}
	time.Sleep(10 * time.Microsecond)
	resp, err := liftingserial.ReadData(nTimeOut)
	if err != nil {
		return "", err
	}
	FDLogger.Printf("cmd:%s--> resp: %s\n", cmd, resp)
	if strings.Contains(resp, "OK\r") {
		return resp, nil
	}
	return "", fmt.Errorf("not found OK: %s", resp)
}

func sendSerialDataATC(cmd string, nTimeOut int32, interval int) (string, error) {
	FDLogger.Printf("\ncmd:%s, timeout:%d, interval:%d\n", cmd, nTimeOut, interval)
	bbcmd := []byte(cmd)
	for i := 0; i < len(bbcmd); i++ {
		if _, err := liftingserial.WriteData(bbcmd[i : i+1]); err != nil {
			return "", err
		}
		time.Sleep(time.Duration(interval) * time.Microsecond)
	}
	time.Sleep(10 * time.Microsecond)
	resp, err := liftingserial.ReadDataATC(nTimeOut)
	if err != nil {
		return "", err
	}
	FDLogger.Printf("cmd:%s--> resp: %s\n", cmd, resp)
	if strings.Contains(resp, "OK\r") {
		return resp, nil
	}
	return "", fmt.Errorf("not found OK: %s", resp)
}

func setNeedSleepFlag() {
	muxWait.Lock()
	bNeedSleep = true
	muxWait.Unlock()
}

func waitSleepFlag() {
	muxWait.Lock()
	time.Sleep(time.Duration(200) * time.Microsecond)
	bNeedSleep = false
	muxWait.Unlock()
}

func hello(c *gin.Context) {
	if bNeedSleep {
		time.Sleep(time.Duration(200) * time.Microsecond)
	}
	liftMutex.Lock()
	defer liftMutex.Unlock()

	defer func() {
		go waitSleepFlag()
	}()
	setNeedSleepFlag()

	cmd := "AT\r"
	sinterval := c.DefaultQuery("interval", "1")
	interval, err := strconv.Atoi(sinterval)
	if err != nil {
		interval = 1
	}
	resp, err := sendSerialData(cmd, 1, interval)
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

func status(c *gin.Context) {
	if bNeedSleep {
		time.Sleep(time.Duration(200) * time.Microsecond)
	}
	liftMutex.Lock()
	defer liftMutex.Unlock()

	defer func() {
		go waitSleepFlag()
	}()
	setNeedSleepFlag()

	cmd := "ATS\r"
	sinterval := c.DefaultQuery("interval", "1")
	interval, err := strconv.Atoi(sinterval)
	if err != nil {
		interval = 1
	}
	resp, err := sendSerialData(cmd, 1, interval)
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

func information(c *gin.Context) {
	if bNeedSleep {
		time.Sleep(time.Duration(200) * time.Microsecond)
	}
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
	resp, err := sendSerialData(cmd, 6, interval)
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
	if bNeedSleep {
		time.Sleep(time.Duration(200) * time.Microsecond)
	}
	liftMutex.Lock()
	defer liftMutex.Unlock()

	defer func() {
		go waitSleepFlag()
	}()
	setNeedSleepFlag()

	cmd := "ATSTOP\r"
	sinterval := c.DefaultQuery("interval", "1")
	interval, err := strconv.Atoi(sinterval)
	if err != nil {
		interval = 1
	}
	resp, err := sendSerialData(cmd, 5, interval)
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

func reset(c *gin.Context) {
	if bNeedSleep {
		time.Sleep(time.Duration(200) * time.Microsecond)
	}
	liftMutex.Lock()
	defer liftMutex.Unlock()

	defer func() {
		go waitSleepFlag()
	}()
	setNeedSleepFlag()

	cmd := "ATZ\r"
	sinterval := c.DefaultQuery("interval", "1")
	interval, err := strconv.Atoi(sinterval)
	if err != nil {
		interval = 1
	}
	resp, err := sendSerialData(cmd, 10, interval)
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

func home(c *gin.Context) {
	if bNeedSleep {
		time.Sleep(time.Duration(200) * time.Microsecond)
	}
	liftMutex.Lock()
	defer liftMutex.Unlock()

	defer func() {
		go waitSleepFlag()
	}()
	setNeedSleepFlag()

	// _, err := sendSerialData("ATG-1\r", 10)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{
	// 		"message": "go to P-1 failed",
	// 		"error":   err.Error(),
	// 	})
	// 	return
	// }

	// time.Sleep(2 * time.Second)

	cmd := "ATC%s\r"
	flag := c.DefaultQuery("flag", "2")
	cmd = fmt.Sprintf(cmd, flag)
	sinterval := c.DefaultQuery("interval", "1")
	interval, err := strconv.Atoi(sinterval)
	if err != nil {
		interval = 1
	}
	resp, err := sendSerialDataATC(cmd, 10, interval)
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

func wind(c *gin.Context) {
	if bNeedSleep {
		time.Sleep(time.Duration(200) * time.Microsecond)
	}
	liftMutex.Lock()
	defer liftMutex.Unlock()

	defer func() {
		go waitSleepFlag()
	}()
	setNeedSleepFlag()

	cmd := "ATW%s\r"
	flag := c.DefaultQuery("flag", "0")
	cmd = fmt.Sprintf(cmd, flag)
	sinterval := c.DefaultQuery("interval", "1")
	interval, err := strconv.Atoi(sinterval)
	if err != nil {
		interval = 1
	}
	resp, err := sendSerialData(cmd, 10, interval)
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

func carrier(c *gin.Context) {
	if bNeedSleep {
		time.Sleep(time.Duration(200) * time.Microsecond)
	}
	liftMutex.Lock()
	defer liftMutex.Unlock()

	defer func() {
		go waitSleepFlag()
	}()
	setNeedSleepFlag()

	cmd := "ATX%s\r"
	flag := c.DefaultQuery("flag", "0")
	cmd = fmt.Sprintf(cmd, flag)
	sinterval := c.DefaultQuery("interval", "1")
	interval, err := strconv.Atoi(sinterval)
	if err != nil {
		interval = 1
	}
	resp, err := sendSerialData(cmd, 10, interval)
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

// PositionInfo position info
type PositionInfo struct {
	Position int    `form:"p"`
	Value    string `form:"value"`
}

func goposition(c *gin.Context) {
	if bNeedSleep {
		time.Sleep(time.Duration(200) * time.Microsecond)
	}
	liftMutex.Lock()
	defer liftMutex.Unlock()

	defer func() {
		go waitSleepFlag()
	}()
	setNeedSleepFlag()

	cmd := "ATG%d\r"
	var pp PositionInfo
	sinterval := c.DefaultQuery("interval", "1")
	interval, err := strconv.Atoi(sinterval)
	if err != nil {
		interval = 1
	}
	if c.ShouldBindQuery(&pp) == nil {
		cmd = fmt.Sprintf("ATG%d\r", pp.Position)
		resp, err := sendSerialData(cmd, 10, interval)
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
		"message": "go posion number",
	})
}

func setPoisition(c *gin.Context) {
	if bNeedSleep {
		time.Sleep(time.Duration(200) * time.Microsecond)
	}
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
		resp, err := sendSerialData(cmd, 3, interval)
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

func listposition(c *gin.Context) {
	if bNeedSleep {
		time.Sleep(time.Duration(200) * time.Microsecond)
	}
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
	resp, err := sendSerialData(cmd, 3, interval)
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
	if bNeedSleep {
		time.Sleep(time.Duration(200) * time.Microsecond)
	}
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
	sinterval := c.DefaultQuery("interval", "1")
	interval, err := strconv.Atoi(sinterval)
	if err != nil {
		interval = 1
	}
	resp, err := sendSerialData(cmd, 10, interval)
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

func flip(c *gin.Context) {
	if bNeedSleep {
		time.Sleep(time.Duration(200) * time.Microsecond)
	}
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
	sinterval := c.DefaultQuery("interval", "1")
	interval, err := strconv.Atoi(sinterval)
	if err != nil {
		interval = 1
	}
	resp, err := sendSerialData(cmd, 10, interval)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": resp,
	})
}
