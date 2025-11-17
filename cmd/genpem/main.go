package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
)

const (
	LengthRSA          = 4096
	PrivateKeyFilePath = `./private.pem`
	PublicKeyFilePath  = `./public.pem`
)

func main() {
	privateKey, err := rsa.GenerateKey(rand.Reader, LengthRSA)
	if err != nil {
		log.Fatal(err)
	}

	fPrivateKeyPEM, err := os.Create(PrivateKeyFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = fPrivateKeyPEM.Close(); err != nil {
			log.Println(err.Error())
		}
	}()
	err = pem.Encode(fPrivateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		log.Fatal(err)
	}

	fPublicKeyPEM, err := os.Create(PublicKeyFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = fPublicKeyPEM.Close(); err != nil {
			log.Println(err.Error())
		}
	}()
	err = pem.Encode(fPublicKeyPEM, &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey),
	})
	if err != nil {
		log.Fatal(err)
	}
}
