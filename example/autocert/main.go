// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package autocert is a simple demo of use smcache with autocert.
package main

import (
	"net/http"

	"github.com/jwendel/smcache"
	"golang.org/x/crypto/acme/autocert"
)

func main() {
	m := &autocert.Manager{
		Cache:      smcache.NewSMCache(smcache.Config{ProjectID: "example-project-1234", SecretPrefix: "test-"}),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("example.com", "www.example.com"),
	}
	s := &http.Server{
		Addr:      ":https",
		TLSConfig: m.TLSConfig(),
	}
	panic(s.ListenAndServeTLS("", ""))
}
