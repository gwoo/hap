// Hap - the simple and effective provisioner
// Copyright (c) 2019 GWoo (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.

package main

import (
	"fmt"
)

// VerboseLogger sets verbose logging either on or off
type VerboseLogger bool

// Print wraps fmt.Print
func (vl VerboseLogger) Print(args ...interface{}) {
	if vl == true {
		fmt.Print(args...)
	}
}

// Println wraps fmt.Println
func (vl VerboseLogger) Println(args ...interface{}) {
	if vl == true {
		fmt.Println(args...)
	}
}

// Printf wraps fmt.Printf
func (vl VerboseLogger) Printf(format string, args ...interface{}) {
	if vl == true {
		fmt.Printf(format, args...)
	}
}
