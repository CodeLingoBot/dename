// Copyright 2014 The Dename Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	_ "crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

func newSerial() *big.Int {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		panic(err)
	}
	return serialNumber
}

func main() {
	if len(os.Args) != 4 {
		log.Fatalf("usage: %s CA_PUBKEY CA_PRIVKEY HOSTIP", os.Args[0])
	}
	ca_cert_file := os.Args[1]
	ca_priv_file := os.Args[2]
	hostip := net.ParseIP(os.Args[3])
	if hostip == nil {
		log.Fatalf("Bad IP: %v", os.Args[3])
	}

	ca_cert_pem, err := ioutil.ReadFile(ca_cert_file)
	if err != nil {
		log.Fatalf("Failed to read from %v: %v", ca_cert_file, err)
	}
	ca_cert_block, _ := pem.Decode(ca_cert_pem)
	if ca_cert_block == nil {
		log.Fatalf("Failed to parse PEM: %v", string(ca_cert_pem))
	}
	ca_cert, err := x509.ParseCertificate(ca_cert_block.Bytes)
	if err != nil {
		log.Fatalf("Failed to parse X.509 cert: %v", err)
	}

	ca_priv_pem, err := ioutil.ReadFile(ca_priv_file)
	if err != nil {
		log.Fatalf("Failed to read from %v: %v", ca_priv_file, err)
	}
	ca_priv_block, _ := pem.Decode(ca_priv_pem)
	if ca_priv_block == nil {
		log.Fatalf("Failed to parse PEM: %v", string(ca_priv_pem))
	}
	ca_priv, err := x509.ParseECPrivateKey(ca_priv_block.Bytes)
	if err != nil {
		log.Fatalf("Failed to parse EC private key: %v", err)
	}

	priv, err := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	if err != nil {
		panic(err)
	}
	cert := &x509.Certificate{
		Subject:      pkix.Name{CommonName: "testingServer"},
		SerialNumber: newSerial(),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(100000 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{hostip},
	}

	der, err := x509.CreateCertificate(rand.Reader, cert, ca_cert, &priv.PublicKey, ca_priv)
	if err != nil {
		panic(err)
	}
	cert, err = x509.ParseCertificate(der)
	if err != nil {
		panic(err)
	}
	err = pem.Encode(os.Stdout, &pem.Block{Type: "CERTIFICATE", Bytes: der})
	if err != nil {
		panic(err)
	}
	test_ca_pool := x509.NewCertPool()
	test_ca_pool.AddCert(ca_cert)
	if _, err := cert.Verify(x509.VerifyOptions{DNSName: "127.0.0.1", Roots: test_ca_pool}); err != nil {
		panic(err)
	}
	skBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		panic(err)
	}
	err = pem.Encode(os.Stderr, &pem.Block{Type: "ECDSA PRIVATE KEY", Bytes: skBytes})
	if err != nil {
		panic(err)
	}
}
