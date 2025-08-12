package main

import (
	"fmt"
	"innoqube-qkd-toolbox/pkg/kmetools"
	"log"
)

func main() {
	var qkdr kmetools.QKDRuntime = kmetools.ArgsValidator()
	status, key := kmetools.KMEKeyGet()
	if status {
		if qkdr.Quiet {
			fmt.Printf("%s:%s\n", key[0], key[1])
		} else {
			log.Printf("[--] KeyID: %s\n", key[0])
			log.Printf("[--] Key: %s\n", key[1])
		}
	} else {
		log.Fatalf("[!!] Error: %s\n", key[1])
	}
}
