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

// secret-test is a sample app that uses smcache without the autocert library.
// It's a good way to validate permissions are setup properly on your GCP account.
package main

import (
	"context"
	"log"

	"github.com/jwendel/smcache"
)

func main() {
	// TODO: you should replace this with your GCP project-id
	projectID := "example-project-1234"

	smc := smcache.NewSMCache(smcache.Config{ProjectID: projectID, SecretPrefix: "testsite-", DebugLogging: true})
	domain := "www.example.com"
	ctx := context.Background()

	// Put some initial data
	err := smc.Put(ctx, domain, []byte("this is some data"))
	if err != nil {
		log.Fatalf("Error on put: %+v", err)
	}

	// Read the data we just stored
	res, err := smc.Get(ctx, domain)
	if err != nil {
		log.Fatalf("failed to get: %v", err)
	}

	log.Printf("got result!:  %+v", string(res))

	// Put a new SecretVersion, to demo a /versions/2 being created
	err = smc.Put(ctx, domain, []byte("new data!"))
	if err != nil {
		log.Fatalf("Error on put: %+v", err)
	}
	// Get the new data we just stored
	res, err = smc.Get(ctx, domain)
	if err != nil {
		log.Fatalf("failed to get: %v", err)
	}

	log.Printf("got result!:  %+v", string(res))

	// Leaving the below commented out so you can see the Secret's in GCP's dashboard.
	// See: https://console.cloud.google.com/security/secret-manager

	// Clean up after ourselves
	// err := smc.Delete(ctx, domain)
	// if err != nil {
	// 	log.Fatalf("failed to delete: %v", err)
	// }
}
