package main

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// Error codes defined in the gopher+ spec
const itemNotFoundErr = 1
const temporaryErr = 2
const itemMovedErr = 3

// I couldn't find a list of "commonly used" error codes, so let's snatch some for ourselves.
const unknownErr = 4

// sendErr sends an error with the specified error code to the client over conn.
// The error is sent using gophertype 3 if gopherP is false, otherwise it's sent the gopher+ way.
// The function prints a log message and returns upon encountering an error while sending the error message.
// It doesn't make sense to pass on the error, as the interaction with the client is already finished (we've answered the request with an error).
func (conn gopherConn) sendErr(msg string, errCode int, gopherP bool) {
	var errStr string
	if gopherP {
		errStr = fmt.Sprintf("--1\r\n%d [%v]\r\n%v\r\n.\r\n", errCode, viper.GetString("admin"), msg)
	} else {
		errStr = fmt.Sprintf("3%v Server admin: [%v].\r\n.", msg, viper.GetString("admin"))
	}

	_, err := conn.Write([]byte(errStr))
	if err != nil {
		log.Printf("Could not send error message \"%v\" to client %v: %v", errStr, conn.RemoteAddr().String(), err.Error())
		return
	}
}

func (conn gopherConn) sendItemNotFoundErr(item string, gopherP bool) {
	conn.sendErr(fmt.Sprintf("The item %v could not be found on this server.", item), itemNotFoundErr, gopherP)
}

func (conn gopherConn) sendTemporaryErr(msg string, gopherP bool) {
	conn.sendErr(fmt.Sprintf("Your request couldn't be served: %v. Please try again later.", msg), temporaryErr, gopherP)
}

func (conn gopherConn) sendItemMovedErr(newSelector string, gopherP bool) {
	conn.sendErr(newSelector, itemMovedErr, gopherP)
}
