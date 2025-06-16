package otto

import "fmt"

var debug bool = false

func SetDebug(b bool) {
	debug = b
}

func Log(v ...any) {
	if debug {
		fmt.Println(v...)
	}
}
