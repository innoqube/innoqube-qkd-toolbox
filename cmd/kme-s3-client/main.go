package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/encrypt"
)

func main() {
	// Minio endpoint and credentials
	endpoint := "localhost:9000" // Replace with your Minio server endpoint
	accessKeyID := "minioadmin"
	secretAccessKey := "minioadmin"
	useSSL := true

	qkd_keyid := "3db4bb3f-0f51-49af-9531-7fb9bc08d1e0"
	qkd_key := "NiXDkmgcAztCFzyhO8XI+COj1Y1pEMDR8H0LzxZxoFo="
	qkd_sae := "CONS_TIM_UPT"

	bucketName := "test-1"
	objectName := "my-secret-data.txt"
	content := "This is a secret message that will be encrypted on the client side only."

	rawseed, err := base64.StdEncoding.DecodeString(qkd_key)
	if err != nil {
		log.Fatalf("[!!] Error decoding Base64 seed: %v", err)
	}

	key := rawseed[:32] // Ensure the key is 32 bytes long for AES-256
	if len(key) != 32 {
		fmt.Println("Error: Key must be 32 bytes long for AES-256")
		os.Exit(1)
	}

	var transport *http.Transport

	if useSSL {
		transport, err = minio.DefaultTransport(useSSL)
		if err != nil {
			fmt.Println("Error creating transport:", err)
			os.Exit(1)
		}
		transport.TLSClientConfig.InsecureSkipVerify = true
	} else {
		transport, err = minio.DefaultTransport(false)
		if err != nil {
			fmt.Println("Error creating transport:", err)
			os.Exit(1)
		}
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure:    useSSL,
		Transport: transport,
	})
	if err != nil {
		fmt.Println("Error initializing Minio client:", err)
		os.Exit(1)
	}

	// Create the bucket if it does not exist
	found, err := client.BucketExists(context.Background(), bucketName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if !found {
		err = client.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("Successfully created bucket:", bucketName)
	}

	// --- Upload the file with client-side encryption ---
	// The key is passed to PutObjectOptions. The Minio SDK uses this key
	// to encrypt the data before it's sent to the server.
	enc, err := encrypt.NewSSEC(key)
	if err != nil {
		fmt.Println("Error creating SSE-C encryptor:", err)
		os.Exit(1)
	}

	objectTags := map[string]string{}
	objectTags["QKD_KeyID"] = qkd_keyid
	objectTags["QKD_SAE"] = qkd_sae

	info, err := client.PutObject(context.Background(), bucketName, objectName, strings.NewReader(content), int64(len(content)), minio.PutObjectOptions{
		ServerSideEncryption: enc,
		UserTags:             objectTags,
	})
	if err != nil {
		fmt.Println("Error uploading encrypted object:", err)
		os.Exit(1)
	}
	fmt.Printf("Successfully uploaded encrypted object: %s with user tags: %s\n", info.Key, objectTags)

	// --- Download the file and decrypt it automatically ---
	// The same key is provided to GetObjectOptions. The Minio SDK uses this key
	// to decrypt the data received from the server.
	// object, err := client.GetObject(context.Background(), bucketName, objectName, minio.GetObjectOptions{
	// 	ServerSideEncryption: minio.NewSSEC(key),
	// })
	// if err != nil {
	// 	fmt.Println("Error downloading encrypted object:", err)
	// 	os.Exit(1)
	// }
	// defer object.Close()

	// Read the decrypted content
	// buf := new(strings.Builder)
	// _, err = strings.NewReader(content).WriteTo(buf)
	// if err != nil {
	// 	fmt.Println("Error reading decrypted object:", err)
	// 	os.Exit(1)
	// }

	// // Read the decrypted object. This will be the original content.
	// readContent := new(strings.Builder)
	// _, err = strings.NewReader(content).WriteTo(readContent)
	// if err != nil {
	// 	fmt.Println("Error reading decrypted object:", err)
	// 	os.Exit(1)
	// }

	// fmt.Printf("Successfully downloaded and decrypted object: %s\n", readContent.String())
}
