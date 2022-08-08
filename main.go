//go:build go1.8
// +build go1.8

package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jxskiss/ginregex"
)

// VERSION is software version
const VERSION = "22.08.08.0"

// var port io.ReadWriteCloser
var (
	powerserial *SerialPort
	FDLogger    *log.Logger
)

// EXITPROC exit processing
var EXITPROC chan string = make(chan string)

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
	c.JSON(http.StatusOK, gin.H{
		"status": "OK",
	})
	EXITPROC <- "server exit"
	FDLogger.Println("Server exiting")
	// os.Exit(0)
}

// HomePage info
func HomePage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version":       VERSION,
		"design":        "jefferyzhang",
		"requestHeader": c.Request.Header,
	})
}

func filterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sPath := c.Request.URL.Path
		if strings.HasPrefix(sPath, "/lift") && liftingserial.portname == "" {
			c.AbortWithStatusJSON(401, gin.H{
				"msg": "not config lift port",
			})
			return
		}
		c.Next()
	}
}

func main() {
	liftport := flag.String("lifter", "", "lifter serial port")
	flag.Parse()
	Init()
	FDLogger.Println("version:" + VERSION)
	FDLogger.Println("http://ip:8010/")
	usbserialList := USBSERIALPORTS{}
	powerserial = &SerialPort{mux: &sync.Mutex{}}
	liftingserial = &ASerialPort{}
	levelserial = &SerialPort{mux: &sync.Mutex{}}

	if usbserialList.LoadConfig("serialcalibration.json") == nil {
		if err := usbserialList.VerifyDevName(); err != nil {
			FDLogger.Fatalf("verifyDevName %s\n", err)
			return
		}
	} else {
		// usbserialList.LoadUSBDevsWithoutConfig()
		usbserialList.LoadUSBDevsWithoutConfigV1()
		powerserial.portname = usbserialList.serialPower
		powerserial.baudrate = usbserialList.PBaudRate
		liftingserial.portname = usbserialList.serialLifting
		liftingserial.baudrate = usbserialList.LBaudRate
		levelserial.portname = usbserialList.serialVoltage
		levelserial.baudrate = usbserialList.LevelBRate
	}

	if usbserialList.serialPower != "" {
		if err := powerserial.Open(usbserialList.serialPower, usbserialList.PBaudRate); err != nil {
			FDLogger.Fatalf("open power control fail: %s\n", err)
			return
		}
		defer powerserial.Close()
	}

	if usbserialList.serialVoltage != "" {
		if err := levelserial.Open(usbserialList.serialVoltage, usbserialList.LevelBRate); err != nil {
			FDLogger.Fatalf("open voltage control fail: %s\n", err)
			return
		}
		defer levelserial.Close()
		sendPowerOn()
	}

	// if usbserialList.serialLifting == "" {
	// 	usbserialList.serialLifting = "/dev/ttyS0"
	// }
	FDLogger.Println("lift port:" + usbserialList.serialLifting)
	if *liftport != "" {
		usbserialList.serialLifting = *liftport
	}
	FDLogger.Println("after lift port:" + usbserialList.serialLifting)
	if usbserialList.serialLifting != "" { // 115200 lifting,
		if err := liftingserial.Open(usbserialList.serialLifting, usbserialList.LBaudRate); err != nil {
			FDLogger.Fatalf("open lift control fail: %s\n", err)
			return
		}
		defer liftingserial.Close()
	}

	go recvStatus()

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(filterMiddleware())
	router.GET("/", HomePage)
	router.POST("/exitsystem", exit)
	v1 := router.Group("/lift")
	{ // for lifter PLC
		v1.GET("/hello", hello)
		v1.GET("/status", status)
		v1.GET("/info", information)
		v1.GET("/position", listposition)
		v1.GET("/go", goposition)
		v1.GET("/home", home)
		v1.GET("/reset", reset)
		v1.GET("/getlasterror", getlasterror)
		v1.GET("/start", start)
		v1.GET("/stop", stop)
		v1.GET("/flip", flip)
		v1.GET("/turn", turn)
		v1.GET("/setpos", setPoisition)
		v1.GET("/wind", wind)
		v1.GET("/carrier", carrier)
		v1.GET("/queryair", queryAirValue)
		v1.GET("/reconnect", func(c *gin.Context) {
			liftingserial.Close()
			FDLogger.Println("Close lifting port")
			err := liftingserial.Open(usbserialList.serialLifting, usbserialList.LBaudRate)
			FDLogger.Println("Open lifting port")
			if err != nil {
				c.JSON(http.StatusForbidden, gin.H{
					"status": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"status": "Success",
			})
		})
	}
	v2 := router.Group("/level")
	{ // for Power Module
		v2.GET("/voltage", voltage)
		v2.GET("/poweron", poweron)
		v2.GET("/poweroff", poweroff)
	}
	v3 := router.Group("/tricl")
	{ // for arduino Module
		v3.GET("/green", greenLed)
		v3.GET("/yellow", yellowLed)
		v3.GET("/red", redLed)
		v3.GET("/clean", cleanLed)
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
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// <-quit
	select {
	case <-quit:
		FDLogger.Println("Shutting down CTRL+C server...")
	case <-EXITPROC:
		FDLogger.Println("http post Shutting down server...")
	}
	//bexit = true
	quitrecv <- true
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
