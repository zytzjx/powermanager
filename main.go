// +build go1.8

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jxskiss/ginregex"
)

// VERSION is software version
const VERSION = "20.09.04.0"

// var port io.ReadWriteCloser
var (
	powerserial *SerialPort
	FDLogger    *log.Logger
)

// Init Loger
func Init() {
	// os.Remove("powermanager.log")
	file, err := os.OpenFile("powermanager.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file:", err)
	}

	multi := io.MultiWriter(file, os.Stdout)
	FDLogger = log.New(multi,
		"",
		log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
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

func exit(c *gin.Context) {
	FDLogger.Println("recv Exit system command")
	type ExitSystem struct {
		Username string `json:"name" binding:"required"`
		Password string `json:"password"`
	}
	var exitsystem ExitSystem
	if err := c.BindJSON(&exitsystem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}
	os.Exit(0)
}

// HomePage info
func HomePage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version":       VERSION,
		"design":        "jefferyzhang",
		"requestHeader": c.Request.Header,
	})
}

func main() {
	Init()
	FDLogger.Println("version:" + VERSION)
	FDLogger.Println("http://ip:8010/")
	usbserialList := USBSERIALPORTS{}
	powerserial = &SerialPort{mux: &sync.Mutex{}}
	liftingserial = &SerialPort{mux: &sync.Mutex{}}

	if usbserialList.LoadConfig("serialcalibration.json") == nil {
		if err := usbserialList.VerifyDevName(); err != nil {
			FDLogger.Fatalf("verifyDevName %s\n", err)
			return
		}
	} else {
		usbserialList.LoadUSBDevsWithoutConfig()
		powerserial.portname = usbserialList.Power
		powerserial.baudrate = usbserialList.PBaudRate
		liftingserial.portname = usbserialList.serialLifting
		liftingserial.baudrate = usbserialList.LBaudRate
	}

	if usbserialList.serialPower != "" {
		if err := powerserial.Open(usbserialList.serialPower, usbserialList.PBaudRate); err != nil {
			FDLogger.Fatalf("open power control fail: %s\n", err)
			return
		}
		defer powerserial.Close()
	}
	if usbserialList.serialLifting != "" { // 115200 lifting,
		if err := liftingserial.Open(usbserialList.serialLifting, usbserialList.LBaudRate); err != nil {
			FDLogger.Fatalf("open power control fail: %s\n", err)
			return
		}
		defer liftingserial.Close()
	}

	router := gin.Default()
	router.GET("/", HomePage)
	router.POST("/exitsystem", exit)
	v1 := router.Group("/lift")
	{
		v1.GET("/hello", hello)
		v1.GET("/status", status)
		v1.GET("/info", information)
		v1.GET("/position", listposition)
		v1.GET("/go", goposition)
		v1.GET("/home", home)
		v1.GET("/reset", reset)
		v1.GET("/stop", stop)
		v1.GET("/flip", flip)
		v1.GET("/turn", turn)
		v1.GET("/setpos", setPoisition)
	}
	regexRouter := ginregex.New(router, nil)
	regexRouter.GET("/\\d+", Power)

	srv := &http.Server{
		Addr:    ":8010",
		Handler: router,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			FDLogger.Fatalf("listen: %s\n", err)
		}
	}()
	// router.Run(":8010")

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	FDLogger.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		FDLogger.Fatalf("Server forced to shutdown: %s\n", err)
	}

	FDLogger.Println("Server exiting")
}
