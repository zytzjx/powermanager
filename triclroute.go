package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

var bexit = false

// var ctx = context.Background()

func recvStatus() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	var re = regexp.MustCompile(`(?m)Key:\s*(\d),\s*(\d),\s*(\d)`)
	// var str = `Key: 0, 5, 2`
	for !bexit {
		time.Sleep(10 * time.Microsecond)
		resp, err := powerserial.ReadDataEnd(1)
		if err != nil && len(resp) < 5 {
			continue
		}
		FDLogger.Printf("cmd resp: %s\n", resp)
		sbstr := re.FindStringSubmatch(resp)
		if len(sbstr) == 4 {
			// Publish a message.
			err = rdb.Publish("tricoloredlight", sbstr[0]).Err()
			if err != nil {
				FDLogger.Println("publish failed:" + err.Error())
			}
			for _, match := range sbstr {
				fmt.Println(match)
			}
			// write redis
			err = rdb.Set("tricoloredlight", sbstr[0], 0).Err()
			if err != nil {
				FDLogger.Println("set tricoloredlight failed:" + err.Error())
			}
		}
	}
}

// Power control power module
func Power(c *gin.Context) {
	FDLogger.Println("power ++")
	param := c.Request.URL.Query()
	re := regexp.MustCompile(`\d+`)
	sp := re.FindString(c.Request.URL.Path)
	dp, err := strconv.Atoi(sp)
	if err != nil || dp < 2 || dp > 13 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": sp,
		})
		FDLogger.Println("power --")
		return
	}
	dp -= 2
	bOn := 1
	if on, ok := param["on"]; ok {
		if len(on) > 0 {
			bn, err := strconv.Atoi(on[0])
			if err != nil {
				log.Fatal(err)
			}
			bOn = bn
		}
	}

	ss := fmt.Sprintf("P%d,%d,\r", dp, bOn)
	FDLogger.Println("Send:", ss)
	if _, err = powerserial.WriteData([]byte(ss)); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"pin":    sp,
			"status": bOn,
			"serial": ss,
		})
		FDLogger.Println("power --")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"pin":    sp,
		"status": bOn,
		"serial": ss,
	})
	FDLogger.Println("power --")
}

func sendcmd(ss string, c *gin.Context) {
	if _, err := powerserial.WriteData([]byte(ss)); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"serial": ss,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"serial": ss,
	})
}

func greenLed(c *gin.Context) {
	sendcmd("g", c)
}

func yellowLed(c *gin.Context) {
	sendcmd("y", c)
}

func redLed(c *gin.Context) {
	sendcmd("r", c)
}

func cleanLed(c *gin.Context) {
	sendcmd("c", c)
}
