package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/spf13/viper"
)

// Variables to identify the build
var (
	Version string
	Commit  string
)

func main() {
	// Load viper config
	setupConfig()
	listener, err := net.Listen("tcp", fmt.Sprintf("%v:%v", viper.Get("address"), viper.Get("port")))
	if err != nil {
		log.Panicf("Error while setting up listener on address %v: %v\n", viper.GetString("addr"), err.Error())
	}
	// Could be empty if built by go build instead of make
	if Version != "" {
		log.Printf("Yagopherd version %v starting...\n", Version)
	} else {
		log.Printf("Yagopherd starting (version unknown)...\n")
	}
	if Commit != "" {
		log.Printf("Commit: %v\n", Commit)
	} else {
		log.Printf("Commit: unknown\n")
	}
	if viper.ConfigFileUsed() != "" {
		log.Printf("Using config file: %s\n", viper.ConfigFileUsed())
	}
	log.Printf("Listening on %v:%v\n", viper.GetString("address"), strconv.Itoa(viper.GetInt("port")))

	// Set up some signal handling stuff to enable a clean shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// This WG is used to allow all handler goroutines to cleanly finish.
	wg := new(sync.WaitGroup)
	// Launch a background "signal monitor" goroutine
	// This isn't done in the main loop because that loop blocks on listener.Accept()
	go func() {
		sig := <-sigs
		log.Printf("Received %v signal, waiting for request processing to finish and other cleanup...\n", sig.String())
		log.Print("Hit CTRL-C again if you wish to terminate without cleanup.")
		// Launch another one if a second SIGTERM is received.
		// This one terminates immediately, without doing any cleanup, useful if the program gets stuck.
		go func() {
			sig := <-sigs
			log.Printf("Received second %v signal, forcing shutdown without cleanup!", sig.String())
			os.Exit(0)
		}()
		wg.Wait()
		// The tests return their own exit code, don't mess with that
		if viper.GetBool("testmode") == false {
			log.Println("Done, shutting down...")
			os.Exit(0)
		}
	}()

	// Main listener loop
	for {
		// TODO: Refuse any new incoming connections upon shutdown init
		rawConn, err := listener.Accept()
		conn := gopherConn{
			rawConn,
		}
		defer conn.Close()

		if err != nil {
			log.Printf("An error occurred while trying to accept request: %v\n", err.Error())
		} else {
			log.Printf("Received request from: %v", net.Addr.String(conn.RemoteAddr()))
			wg.Add(1)
			go handleReq(conn, wg)
		}
	}
}

// handleReq handles an incoming request by parsing the selector and sending the selected content to the client.
func handleReq(conn gopherConn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()

	// Extract attributes of the request
	req, err := extractReq(conn)
	if err != nil {
		// As the function failed it's unknown whether client supports gopher+, assume gopher
		conn.sendErr(err.Error(), unknownErr, false)
	}

	// Blank selector = request for gopherroot
	if req.selector == "" {
		req.path = viper.GetString("gopherroot")
	} else {
		req.path, err = appendDir(viper.GetString("gopherroot"), req.selector)
		if err != nil {
			conn.sendErr(err.Error(), unknownErr, req.gopherP)
			return
		}
	}

	fInfo, err := os.Stat(req.path)
	if err != nil {
		// TODO: Find out syscall error codes and only send itemNotFoundError when appropriate
		conn.sendItemNotFoundErr(req.selector, req.gopherP)
		return
	}

	if fInfo.IsDir() == true {
		gophermap, err := dirToGophermap(req.path, req.gopherP)
		if err != nil {
			if _, ok := err.(*os.PathError); ok {
				conn.sendItemNotFoundErr(req.path, req.gopherP)
			} else {
				conn.sendErr(err.Error(), unknownErr, req.gopherP)
			}
			return
		}
		err = conn.sendGophermap(gophermap)
		if err != nil {
			conn.sendErr(err.Error(), unknownErr, req.gopherP)
			return
		}
	} else {
		err = conn.sendFile(req.path, req.gopherP)
		if err != nil {
			conn.sendErr(err.Error(), unknownErr, req.gopherP)
			return
		}
	}
	return
}
