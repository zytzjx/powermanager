// +build go1.8

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jacobsa/go-serial/serial"
)

func main() {

	fmt.Println("http://ip:8010/[pinnumber]?on=[0|1]")
	// Set up options.
	options := serial.OpenOptions{
		PortName:        "/dev/ttyUSB0",
		BaudRate:        19200,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}

	// Open the port.
	port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}

	// Make sure to close it later.
	defer port.Close()

	router := gin.Default()
	router.GET("/:pin", func(c *gin.Context) {
		param := c.Request.URL.Query()
		sp := c.Param("pin")
		dp, err := strconv.Atoi(sp)
		if err != nil || dp < 2 || dp > 13 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": sp,
			})
		}
		dp -= 2
		bOn := 1
		if on, ok := param["on"]; ok {
			if len(on) > 0 {
				if bn, err := strconv.Atoi(on[0]); err == nil {
					bOn = bn
				}
			}
		}

		ss := fmt.Sprintf("P%d,%d", dp, bOn)
		port.Write([]byte(ss))
		c.JSON(http.StatusOK, gin.H{
			"pin":    sp,
			"status": bOn,
			"serial": ss,
		})
	})

	srv := &http.Server{
		Addr:    ":8010",
		Handler: router,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
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
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
