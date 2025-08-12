// SPDX-License-Identifier: Zlib
// Copyright 2025, Terry M. Poulin.
package logging

import (
	"fmt"
	"time"
)

// Calls the provided function to log a message containing the provided string
// and the start/finish information. Returns a function that can be called to
// log the elapsed time.
func When(str string, log func(...any)) func() {
	now := time.Now()
	log(fmt.Sprintf("Started %s at %s", str, now))
	return func() {
		finished := time.Now()
		log(fmt.Sprintf("Finished %s at %s (%s elapsed)", str, finished, now.Sub(finished)))
	}
}
