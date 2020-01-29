GSMCache is an implementation of the [Cache](https://godoc.org/golang.org/x/crypto/acme/autocert#Cache) within [acme autocert](https://godoc.org/golang.org/x/crypto/acme/autocert) that will store data within [Google Cloud's Secret Manager](https://cloud.google.com/secret-manager/docs).

NOTE: This is a work-in-progress, API is NOT stable. It currently works with one of my sites, there are no unit tests yet and I haven't finished fleshing out any corner cases.

# How To Use

```go

import (
	"github.com/jwendel/gsmcache"
	"golang.org/x/crypto/acme/autocert"
)

...

m := &autocert.Manager{
    Cache:      &gsmcache.GSMCache{ProjectId: "my-project-id", SecretPrefix: "test-"},
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

- Tests
  - Validate empty SecretPrefix works.
- Flag for debug logging.
- Flag to not delete SecretVersion on update.
