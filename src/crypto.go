package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"math/big"
	"strings"
)

func checkKeyPair() (privateKey string, publicKey string) {
	// check if key pair exists
	privateKey, publicKey = getKeyPair()
	if privateKey != "" && publicKey != "" {
		return
	}
	// generate key pair
	privateKey, publicKey, err := generateKeyPair()
	fatalErr(err)
	insertKeyPair(privateKey, publicKey)
	return
}

func generateKeyPair() (privateKey string, publicKey string, e error) {
	priKey, e := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if e != nil {
		return "", "", e
	}
	ecPrivateKey, e := x509.MarshalECPrivateKey(priKey)
	if e != nil {
		return "", "", e
	}
	privateKey = base64.StdEncoding.EncodeToString(ecPrivateKey)

	X := priKey.X
	Y := priKey.Y
	xStr, e := X.MarshalText()
	if e != nil {
		return "", "", e
	}
	yStr, e := Y.MarshalText()
	if e != nil {
		return "", "", e
	}
	public := string(xStr) + "+" + string(yStr)
	publicKey = base64.StdEncoding.EncodeToString([]byte(public))
	return
}

func getPrivateKey(privateKeyStr string) (priKey *ecdsa.PrivateKey, e error) {
	bytes, e := base64.StdEncoding.DecodeString(privateKeyStr)
	if e != nil {
		return nil, e
	}
	priKey, e = x509.ParseECPrivateKey(bytes)
	if e != nil {
		return nil, e
	}
	return
}

func getPublicKey(publicKeyStr string) (pubKey *ecdsa.PublicKey, e error) {
	bytes, e := base64.StdEncoding.DecodeString(publicKeyStr)
	if e != nil {
		return nil, e
	}
	split := strings.Split(string(bytes), "+")
	xStr := split[0]
	yStr := split[1]
	x := new(big.Int)
	y := new(big.Int)
	e = x.UnmarshalText([]byte(xStr))
	if e != nil {
		return nil, e
	}
	e = y.UnmarshalText([]byte(yStr))
	if e != nil {
		return nil, e
	}
	pub := ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}
	pubKey = &pub
	return
}

func sign(data string, privateKeyStr string) (string, error) {
	privateKey, err := getPrivateKey(privateKeyStr)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256([]byte(data))
	sign, err := ecdsa.SignASN1(rand.Reader, privateKey, hash[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sign), nil
}

func verify(data string, publicKeyStr string, sign string) (bool, error) {
	publicKey, err := getPublicKey(publicKeyStr)
	if err != nil {
		return false, err
	}
	hash := sha256.Sum256([]byte(data))
	signb, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false, err
	}
	return ecdsa.VerifyASN1(publicKey, hash[:], signb), nil
}
