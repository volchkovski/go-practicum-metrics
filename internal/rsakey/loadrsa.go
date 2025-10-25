package rsakey

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

var ErrEmptyPath = fmt.Errorf("empty path")

func GetPublicKey(filePath string) (*rsa.PublicKey, error) {
	if filePath == "" {
		return nil, ErrEmptyPath
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	return x509.ParsePKCS1PublicKey(block.Bytes)
}

func GetPrivateKey(filePath string) (*rsa.PrivateKey, error) {
	if filePath == "" {
		return nil, ErrEmptyPath
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
