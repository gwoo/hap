// Hap - the simple and effective provisioner
// Copyright (c) 2014 Garrett Woodworth (https://github.com/gwoo)
// The BSD License http://opensource.org/licenses/bsd-license.php.

package main

import (
	"fmt"
)

type VerboseLogger bool

func (vl VerboseLogger) Print(args ...interface{}) {
	if vl == true {
		fmt.Print(args...)
	}
}

func (vl VerboseLogger) Println(args ...interface{}) {
	if vl == true {
		fmt.Println(args...)
	}
}
func (vl VerboseLogger) Printf(format string, args ...interface{}) {
	if vl == true {
		fmt.Printf(format, args...)
	}
}
