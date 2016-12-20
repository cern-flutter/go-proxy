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

package main

import (
	"crypto/rsa"
	"flag"
	"fmt"
	"gitlab.cern.ch/flutter/go-proxy"
	"log"
	"os"
	"time"
)

var proxyPath = flag.String("file", "", "Proxy location")
var typeRepr = []string{"Not a proxy", "Legacy proxy", "Draft proxy", "RFC 3820 Proxy"}

func getProxyPath() string {
	if path := os.Getenv("X509_USER_PROXY"); path != "" {
		return path
	}

	return fmt.Sprintf("/tmp/x509up_u%d", os.Getuid())
}

func main() {
	capath := flag.String("capath", "/etc/grid-security/certificates", "Directory with the root CAs")
	vomsdir := flag.String("vomsdir", "/etc/grid-security/vomsdir/", "VOMS dir")

	flag.Parse()

	if *proxyPath == "" {
		*proxyPath = getProxyPath()
	}

	var p proxy.X509Proxy
	if e := p.DecodeFromFile(*proxyPath); e != nil {
		log.Fatal(e)
	}

	fmt.Printf("subject   : %s\n", proxy.NameRepr(p.Subject))
	fmt.Printf("issuer    : %s\n", proxy.NameRepr(p.Issuer))
	fmt.Printf("identity  : %s\n", proxy.NameRepr(p.Identity))
	fmt.Printf("type      : %s\n", typeRepr[p.ProxyType])
	fmt.Printf("strength  : %d bits\n", p.Certificate.PublicKey.(*rsa.PublicKey).N.BitLen())
	fmt.Printf("timeleft  : %s\n", p.Certificate.NotAfter.Sub(time.Now()))
	fmt.Printf("key usage : %s\n", proxy.KeyUsageRepr(p.Certificate.KeyUsage))
	if len(p.VomsAttributes) > 0 {
		fmt.Print("=== VO dteam extension information ===\n")
	}
	for _, v := range p.VomsAttributes {
		fmt.Printf("VO        : %s\n", v.Vo)
		fmt.Printf("subject   : %s\n", proxy.NameRepr(v.Subject))
		fmt.Printf("issuer    : %s\n", proxy.NameRepr(v.Issuer))
		fmt.Printf("attribute : %s\n", v.Fqan)
		fmt.Printf("timeleft  : %s\n", v.NotAfter.Sub(time.Now()))
		fmt.Printf("uri       : %s\n", v.PolicyAuthority)
	}

	if roots, err := proxy.LoadCAPath(*capath); err != nil {
		fmt.Printf("Failed to load the root CA: %s", err)
	} else if err = p.Verify(proxy.VerifyOptions{
		Roots:   roots,
		VomsDir: *vomsdir,
	}); err != nil {
		fmt.Printf("Verification result: %s\n", err)
	} else {
		fmt.Printf("Verification OK\n")
	}
}
