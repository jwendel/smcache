SMCache is an implementation of the [Cache](https://godoc.org/golang.org/x/crypto/acme/autocert#Cache) within [acme autocert](https://godoc.org/golang.org/x/crypto/acme/autocert) that will store data within [Google Cloud's Secret Manager](https://cloud.google.com/secret-manager/docs).

**This is not an official Google product.**

NOTE: This is a work-in-progress, API is NOT stable. It seems to work, but not all corner-cases have been tested yet.

# How To Use

```go

import (
	"github.com/google/smcache"
	"golang.org/x/crypto/acme/autocert"
)

...

m := &autocert.Manager{
    Cache:      &smcache.SMCache{ProjectId: "my-project-id", SecretPrefix: "test-"},
    Prompt:     autocert.AcceptTOS,
    HostPolicy: autocert.HostWhitelist("example.com", "www.example.com"),
}
s := &http.Server{
    Addr:         ":https",
    TLSConfig:    m.TLSConfig(),
}

panic(s.ListenAndServeTLS("", ""))
```

# Dev TODO List

- [ ] Tests
  - [X] mocks created, basic test works.
  - [X] Validate unset SecretPrefix works.
  - [X] Get tests
  - [ ] Put tests
  - [ ] Delete tests
- [X] Flag for debug logging.
- [ ] Flag to not delete SecretVersion on update.
