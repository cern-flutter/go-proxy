/*
 * Copyright (c) CERN 2016
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package proxy

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestParseMalformed(t *testing.T) {
	var p X509Proxy
	if e := p.Decode([]byte("RANDOM GARBAGE")); e == nil {
		t.Fatal("Expected an error")
	}
}

func commonAsserts(proxy *X509Proxy, t *testing.T) {
	if proxy == nil {
		t.Fatal("Proxy must not be nil")
	}
	if proxy.Raw == nil {
		t.Fatal("No certificate loaded")
	}
	if proxy.PrivateKey != nil {
		t.Fatal("Unexpected private key")
	}
	if NameRepr(&proxy.Identity) != "/DC=ch/DC=cern/OU=Organic Units/OU=Users/CN=saketag/CN=678984/CN=Alejandro Alvarez Ayllon" {
		t.Fatal("Unexpected user dn: ", proxy.Subject)
	}
	if len(proxy.VomsAttributes) != 1 {
		t.Fatal("Was expecting at least one VO extension")
	}
	if proxy.VomsAttributes[0].Vo != "dteam" {
		t.Fatal("Expecting dteam VO")
	}
	if proxy.VomsAttributes[0].PolicyAuthority != "dteam://voms2.hellasgrid.gr:15004" {
		t.Fatal("Unexpected PolicyAuthority")
	}
	if len(proxy.Chain) != 1 {
		t.Fatal("Expecting one certificate in the chain")
	}
	if proxy.DelegationID() == "" {
		t.Fatal("Delegation id empty")
	}
}

func loadProxy(path string, t *testing.T) *X509Proxy {
	content, e := ioutil.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	p := &X509Proxy{}
	if e = p.Decode(content); e != nil {
		t.Fatal(e)
	}
	return p
}

func TestLegacyProxy(t *testing.T) {
	p := loadProxy("test-samples/LegacyProxy.pem", t)
	commonAsserts(p, t)
	if p.ProxyType != TypeLegacy {
		t.Fatal("Expecting Legacy proxy")
	}
}

func TestDraftProxy(t *testing.T) {
	p := loadProxy("test-samples/DraftProxy.pem", t)
	commonAsserts(p, t)
	if p.ProxyType != TypeDraft {
		t.Fatal("Expecting Draft proxy")
	}
}

func TestRfcProxy(t *testing.T) {
	p := loadProxy("test-samples/RfcProxy.pem", t)
	commonAsserts(p, t)
	if p.ProxyType != TypeRFC3820 {
		t.Fatal("Expecting RFC proxy")
	}
}

func TestSerialize(t *testing.T) {
	original, e := ioutil.ReadFile("test-samples/RfcProxy.pem")
	if e != nil {
		t.Fatal(e)
	}
	p := &X509Proxy{}
	if e = p.Decode(original); e != nil {
		t.Fatal(e)
	}
	// loadChain appends an extra \n!
	pem := p.Encode()
	if !bytes.Equal(original[:len(original)], pem) {
		t.Fatal("Serialized version does not match the original one")
	}
}
