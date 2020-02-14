# Overview

Go library to store certificates from Let's Encrypt in GCP Secret Manager.

SMCache is an implementation of the [Cache](https://godoc.org/golang.org/x/crypto/acme/autocert#Cache) within [acme autocert](https://godoc.org/golang.org/x/crypto/acme/autocert) that will store data within [Google Cloud's Secret Manager](https://cloud.google.com/secret-manager/docs).

**This is not an official Google product.**

NOTE: This is a work-in-progress, API is NOT stable. It seems to work, but not all corner-cases have been tested yet.

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

SMCache requires [admin access](https://cloud.google.com/secret-manager/docs/access-control) to the secret Manager API to function properly. This is configure in the IAM policy for a resource. 

Example of enabling this API for Compute Engine:

1) Go the [IAM policy management](https://console.cloud.google.com/iam-admin/iam)
2) Edit the `<projectId>-compute@developer.gserviceaccount.com` (`Compute Engine default service account`)
3) Click `Add Another Role`, and select `Secret Manager Admin`.

Bonus Security: if you're paranoid about this resource getting access to other secrets, you can set a condition on the Role we just added.

4) click `Add Condition`, then set a name and description for it.
5) For Contidition Type, select `Resource` -> `Name`, Operator: `Starts With`, and set it to whatever value you want, such as "`test-`".
   * this prefix likely should be the same as the `SecretPrefix` you set on the `smcache.Config`.

# Dev TODO List

- [ ] Tests
  - [X] mocks created, basic test works.
  - [X] Validate unset SecretPrefix works.
  - [X] Get tests
  - [ ] Put tests (started)
  - [ ] Delete tests
- [X] Flag for debug logging.
- [ ] Flag to not delete SecretVersion on update.
