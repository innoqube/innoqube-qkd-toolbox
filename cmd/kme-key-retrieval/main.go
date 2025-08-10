package main

import (
	"fmt"
	"innoqube-qkd-toolbox/pkg/iqutils"
	"log"
)

func main() {
	var qkdr iqutils.QKDRuntime = iqutils.ArgsValidator()
	status, key := iqutils.KMEKeyGet()
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
