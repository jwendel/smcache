# Overview

[![GoDoc](https://godoc.org/github.com/jwendel/smcache?status.svg)](https://godoc.org/github.com/jwendel/smcache)
[![Build Status](https://api.travis-ci.org/jwendel/smcache.svg?branch=master&label=Windows+and+Linux+and+macOS+build "Build Status")](https://travis-ci.org/jwendel/smcache)
[![Go Report Card](https://goreportcard.com/badge/github.com/jwendel/smcache)](https://goreportcard.com/report/github.com/jwendel/smcache)

SMCache is a Go library to store certificates from Let's Encrypt in GCP Secret Manager.
It is an implementation of the [Cache](https://godoc.org/golang.org/x/crypto/acme/autocert#Cache)
within [acme autocert](https://godoc.org/golang.org/x/crypto/acme/autocert) 
that will store data within [Google Cloud's Secret Manager](https://cloud.google.com/secret-manager/docs).

> This is not an official Google product.

# Simple Example

```go
import (
	"github.com/jwendel/smcache"
	"golang.org/x/crypto/acme/autocert"
)

func main() {
  m := &autocert.Manager{
      Cache:      smcache.NewSMCache(smcache.Config{ProjectID: "my-project-id", SecretPrefix: "test-"}),
      Prompt:     autocert.AcceptTOS,
      HostPolicy: autocert.HostWhitelist("example.com", "www.example.com"),
  }
  s := &http.Server{
      Addr:         ":https",
      TLSConfig:    m.TLSConfig(),
  }
  panic(s.ListenAndServeTLS("", ""))
}
```

# Detailed Guide to Setting up SMCache

## Permission setup in GCP

SMCache requires [admin access](https://cloud.google.com/secret-manager/docs/access-control) to the Secret Manager API to function properly. This is configure in the IAM policy for a resource. 

Example of enabling this API for Compute Engine:

1) Go the [IAM policy management](https://console.cloud.google.com/iam-admin/iam)
2) Edit the `<projectId>-compute@developer.gserviceaccount.com` (`Compute Engine default service account`)
3) Click `Add Another Role`, and select `Secret Manager Admin`.

Bonus Security: if you're paranoid about this resource getting access to other secrets, you can set a condition on the Role we just added.

4) click `Add Condition`, then set a name and description for it.
5) For Conditional Type, select `Resource` -> `Name`, Operator: `Starts With`, and set it to whatever value you want, such as "`test-`".
   * Note: this prefix should be the same as the `SecretPrefix` you set on the `smcache.Config`.

## Demos

There are 2 demos checked into this repo under example/.

* [Autocert+Http Example](https://github.com/jwendel/smcache/tree/master/example/autocert) - shows how to use this library with Autocert and the Go HTTP std server.
* [Simple Example](https://github.com/jwendel/smcache/tree/master/example/simple) - demos how this library interacts with GCP's Secret Manager.

## Other notes

* Requires Go >= 1.13.0 (due to use of `fmt.Errorf`)
