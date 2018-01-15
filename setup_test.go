package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/phayes/freeport"
	"github.com/spf13/viper"
)

// Addr is the Address of the server under test.
// It's a global variable so that all test cases can access it easily.
var Addr *net.TCPAddr

// Start the server on a random high port assigned by the kernel
func TestMain(m *testing.M) {
	// Tell yagopherd to not run stuff that would interfere with the tests
	viper.Set("testmode", true)

	var err error
	viper.Set("gopherroot", "./testdata/gopherroot")
	port, err := freeport.GetFreePort()
	if err != nil {
		log.Fatal(err)
	}
	viper.Set("port", port)
	viper.Set("address", "localhost")

	// Call main() to start the server
	// A goroutine is used so the tests can be run at the same time
	go main()

	// Resolve the server's address for use in test cases.
	Addr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%v:%v", viper.GetString("address"), strconv.Itoa(viper.GetInt("port"))))
	if err != nil {
		log.Fatal(err)
	}
	// Wait for the server to start, deadline is 5 seconds.
	for i := 0; i < 5; i++ {
		_, err = net.DialTCP("tcp", nil, Addr)
		if i > 4 {
			log.Fatal("Server bringup timed out.")
		}
		// Error -> server is not up and listening, retry
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		// No error -> succesfull connection -> server is running
		if err == nil {
			break
		}
	}
	retCode := m.Run()

	// Do some teardown
	// os.Exit() does not send a signal to the server process, which means it won't terminate cleanly.
	// Therefore we signal ourselves, invoking the shutdown handler.
	process, err := os.FindProcess(os.Getpid())
	if err != nil {
		log.Fatal(err)
	}
	process.Signal(os.Interrupt)
	// The shutdown handler has no way to know whether the tests were successful, therefore we do our own os.Exit() with our own return code.
	os.Exit(retCode)
}
