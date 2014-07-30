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
package client

import (
	"code.google.com/p/goprotobuf/proto"
	"crypto/rand"
	. "github.com/andres-erbsen/dename/protocol"
	"fmt"
	"github.com/agl/ed25519"
	"io"
	"time"
)

var fieldByName = map[string]int32{
	"ssh-host": 21,    // a line for .ssh/known_hosts
	"ssh":      22,    // first to two space-separated fields for .ssh/authorzed_keys
	"email":    25,    // an email address
	"dns":      25,    // a domain name
	"http":     80,    // a http:// or https:// url
	"web":      80,    // ^
	"xmpp":     5222,  // a XMPP address
	"jabber":   5222,  // ^
	"otr":      5223,  // 20 bytes: OTR fingerprint
	"pgp":      11371, // 20 bytes: an OpenPGP key fingerprint
	"gpg":      11371, // ^
	"openpgp":  11371, // ^
}

func FieldByName(fieldName string) (fieldNumber int32, err error) {
	if fieldNumber, ok := fieldByName[fieldName]; ok {
		return fieldNumber, nil
	}
	_, err = fmt.Sscan(fieldName, &fieldNumber)
	return
}

func profileField(field int32) (ret *proto.ExtensionDesc) {
	ret = &proto.ExtensionDesc{(*Profile)(nil), ([]byte)(nil),
		field, fmt.Sprint(field), "bytes," + fmt.Sprint(field) + ",opt",
	}
	func() {
		defer func() { recover() }() // panic if already registered
		proto.RegisterExtension(ret)
	}()
	return
}

func GetProfileField(profile *Profile, field int32) ([]byte, error) {
	ret, err := proto.GetExtension(profile, profileField(field))
	if err != nil {
		return nil, err
	}
	return ret.([]byte), err
}

func SetProfileField(profile *Profile, field int32, value []byte) error {
	return proto.SetExtension(profile, profileField(field), value)
}

func NewProfile(rnd io.Reader, expirationTime *time.Time) (profile *Profile,
	sk *[ed25519.PrivateKeySize]byte, err error) {
	if rnd == nil {
		rnd = rand.Reader
	}
	pk, sk, err := ed25519.GenerateKey(rnd)
	if expirationTime == nil {
		t := time.Now().Add(MAX_VALIDITY_PERIOD*time.Second - 24*time.Hour)
		expirationTime = &t
	}
	et := uint64(expirationTime.Unix())
	profile = &Profile{
		SignatureKey:   &Profile_PublicKey{Ed25519: pk[:]},
		ExpirationTime: &et,
	}
	return
}