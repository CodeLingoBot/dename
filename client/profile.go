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
	"crypto/rand"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/agl/ed25519"
	. "github.com/andres-erbsen/dename/protocol"
	"github.com/gogo/protobuf/proto"
)

var fieldByName = map[string]int32{
	"ssh-host":         21,    // lines for .ssh/known_hosts
	"ssh":              22,    // first two space-separated fields for .ssh/authorzed_keys
	"email":            25,    // email address
	"dns":              53,    // domain name
	"http":             80,    // http:// or https:// url
	"web":              80,    // ^
	"xmpp":             5222,  // XMPP address
	"jabber":           5222,  // ^
	"otr":              5223,  // 40 bytes: hex-encoded OTR fingerprint without spaces
	"dename-transport": 6263,  // base64-encoded 32-byte Curve25519 public key
	"bitcoin":          8333,  // bitcoin id
	"dename":           8877,  // dename server signature key: base64-encoded Profile.PublicKey
	"tor":              9050,  // 40 bytes: hex-encoded fingerprint of a tor node
	"pgp":              11371, // 40 bytes: hex-encoded OpenPGP key fingerprint without spaces
	"gpg":              11371, // ^
	"openpgp":          11371, // ^
	"textsecure":       13785, // 66 bytes: hex encoded fingerprint of a textsecure user
}

func FieldByName(fieldName string) (fieldNumber int32, err error) {
	if fieldNumber, ok := fieldByName[fieldName]; ok {
		return fieldNumber, nil
	}
	_, err = fmt.Sscan(fieldName, &fieldNumber)
	return
}

var profileFieldDescr = map[int32]*proto.ExtensionDesc{}
var profileFieldDescrMu sync.Mutex

func profileField(field int32) *proto.ExtensionDesc {
	profileFieldDescrMu.Lock()
	defer profileFieldDescrMu.Unlock()
	if descr, ok := profileFieldDescr[field]; ok {
		return descr
	}
	descr := &proto.ExtensionDesc{(*Profile)(nil), ([]byte)(nil),
		field, fmt.Sprint(field), "bytes," + fmt.Sprint(field) + ",opt",
	}
	proto.RegisterExtension(descr)
	profileFieldDescr[field] = descr
	return descr
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

func ClearProfileField(profile *Profile, field int32) {
	proto.ClearExtension(profile, profileField(field))
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
