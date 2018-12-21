package logger

import "fmt"

// Logger is responsible to log output to the user
type logger struct {
	debug bool
}

// Logger is an accessible singleton to provide logging
var log logger

func Debug(b bool) {
	log.debug = b
}

func Log(msg string) {
	fmt.Println(msg)
}

func Error(err error, msg string) {
	fmt.Println("ERROR: " + msg)
	if log.debug && err != nil {
		fmt.Printf("err: %s\n", err.Error())
	}
}
