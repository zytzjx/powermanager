package main

import (
	"bufio"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

var liftingserial *SerialPort

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

func sendSerialData(cmd string, nTimeOut int32) (string, error) {
	FDLogger.Printf("cmd:%s, timeout:%d\n", cmd, nTimeOut)
	if _, err := liftingserial.WriteData([]byte(cmd)); err != nil {
		return "", err
	}
	resp, err := liftingserial.ReadData(nTimeOut)
	if err != nil {
		return "", err
	}
	FDLogger.Printf("cmd:%s--> resp: %s\n", cmd, resp)
	if strings.Contains(resp, "OK\r\n") {
		return resp, nil
	}
	return "", fmt.Errorf("not found OK: %s", resp)
}

func hello(c *gin.Context) {
	cmd := "AT\r"
	resp, err := sendSerialData(cmd, 1)
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

func status(c *gin.Context) {
	cmd := "ATS\r"
	resp, err := sendSerialData(cmd, 1)
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

func information(c *gin.Context) {
	cmd := "ATI\r"
	resp, err := sendSerialData(cmd, 6)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
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
	cmd := "ATSTOP\r"
	resp, err := sendSerialData(cmd, 5)
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

func reset(c *gin.Context) {
	cmd := "ATZ\r"
	resp, err := sendSerialData(cmd, 10)
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

func home(c *gin.Context) {
	cmd := "ATC%s\r"
	flag := c.DefaultQuery("flag", "1")
	cmd = fmt.Sprintf(cmd, flag)
	resp, err := sendSerialData(cmd, 10)
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

// PositionInfo position info
type PositionInfo struct {
	Position int    `form:"p"`
	Value    string `form:"value"`
}

func goposition(c *gin.Context) {
	cmd := "ATG%d\r"
	var pp PositionInfo
	if c.ShouldBindQuery(&pp) == nil {
		cmd = fmt.Sprintf("ATG%d\r", pp.Position)
		resp, err := sendSerialData(cmd, 10)
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
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{
		"status":  "Error",
		"message": "go posion number",
	})
}

func setPoisition(c *gin.Context) {
	var pp PositionInfo
	if c.ShouldBindQuery(&pp) == nil {
		val, err := strconv.ParseFloat(pp.Value, 64)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err,
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
		resp, err := sendSerialData(cmd, 1)
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
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{
		"status":  "Error",
		"message": "set posion number & value",
	})
}

func listposition(c *gin.Context) {
	cmd := "ATP?\r"
	var pp PositionInfo
	if c.ShouldBindQuery(&pp) == nil {
		cmd = fmt.Sprintf("ATP%d?\r", pp.Position)
	}
	resp, err := sendSerialData(cmd, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
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
	_, err := sendSerialData("ATG-1\r", 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "go to P-1 failed",
			"error":   err,
		})
		return
	}
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
	resp, err := sendSerialData(cmd, 10)
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

func flip(c *gin.Context) {
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
	resp, err := sendSerialData(cmd, 10)
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
