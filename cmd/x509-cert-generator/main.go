package main

import (
	"encoding/pem"
	"innoqube-qkd-toolbox/pkg/kmetools"
	"log"
	"os"
)

func main() {
	var qkdr kmetools.QKDRuntime = kmetools.ArgsValidator()
	privKeyBytes, certBytes := kmetools.KMEx509Generator()

	if qkdr.X509PrintOut {
		pem.Encode(os.Stdout, &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: privKeyBytes,
		})
		pem.Encode(os.Stdout, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certBytes,
		})
	} else {
		privKeyFile, err := os.Create(qkdr.X509FilePrefix + "-privatekey.pem")
		if err != nil {
			log.Fatalf("Failed to open %s-privatekey.pem: %v\n", qkdr.X509FilePrefix, err)
		}
		defer privKeyFile.Close()

		pem.Encode(privKeyFile, &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: privKeyBytes,
		})
		log.Printf("Private key saved to %s-privatekey.pem", qkdr.X509FilePrefix)
		certFile, err := os.Create(qkdr.X509FilePrefix + "-certificate.pem")
		if err != nil {
			log.Fatalf("Failed to open %s-certificate.pem: %v\n", qkdr.X509FilePrefix, err)
		}
		defer certFile.Close()

		pem.Encode(certFile, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certBytes,
		})
		log.Printf("Certificate saved to %s-certificate.pem", qkdr.X509FilePrefix)
	}
}
