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
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// var port io.ReadWriteCloser
var (
	powerserial SerialPort
	FDLogger    *log.Logger
)

// Init Loger
func Init() {
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
	param := c.Request.URL.Query()
	sp := c.Param("pin")
	dp, err := strconv.Atoi(sp)
	if err != nil || dp < 2 || dp > 13 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": sp,
		})
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

	ss := fmt.Sprintf("P%d,%d", dp, bOn)
	FDLogger.Println("Send:", ss)
	if _, err = powerserial.WriteData([]byte(ss)); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"pin":    sp,
			"status": bOn,
			"serial": ss,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"pin":    sp,
		"status": bOn,
		"serial": ss,
	})
}

func exit(c *gin.Context) {
	FDLogger.Println("recv Exit system command")
	os.Exit(0)
}

func main() {
	Init()
	FDLogger.Println("version:20.08.06.0")
	FDLogger.Println("http://ip:8010/")
	usbserialList := USBSERIALPORTS{}
	usbserialList.LoadConfig("calibration.json")
	if err := usbserialList.verifyDevName(); err != nil {
		FDLogger.Fatalf("verifyDevName %s\n", err)
		return
	}

	powerserial := SerialPort{}
	liftingserial := SerialPort{}
	if err := powerserial.Open(usbserialList.serialPower, 9600); err != nil {
		FDLogger.Fatalf("open power control fail: %s\n", err)
		return
	}
	defer powerserial.Close()

	if err := liftingserial.Open(usbserialList.serialLifting, 115200); err != nil {
		FDLogger.Fatalf("open power control fail: %s\n", err)
		return
	}
	defer liftingserial.Close()

	// // Set up options.
	// options := serial.OpenOptions{
	// 	PortName:        "/dev/ttyUSB0",
	// 	BaudRate:        9600,
	// 	DataBits:        8,
	// 	StopBits:        1,
	// 	MinimumReadSize: 4,
	// }

	// // Open the port.
	// port, err := serial.Open(options)
	// if err != nil {
	// 	log.Fatalf("serial.Open: %v", err)
	// }

	// // Make sure to close it later.
	// defer port.Close()

	router := gin.Default()
	router.GET("/:pin", Power)
	router.GET("/exitsystem", exit)
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
