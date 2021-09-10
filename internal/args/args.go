/* Copyright (c) 2021 Eric St-Amand
See LICENSE for details. */

package args

import (
	"flag"
	"log"
)

func Parse() string {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		log.Fatal("pgsanity: missing valid file or directory argument")
	}
	if len(args) > 1 {
		log.Fatal("pgsanity: too many arguments, only 1 argument supported")
	}
	return args[0]
}
