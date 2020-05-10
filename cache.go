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

package smcache

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/jwendel/smcache/internal/api"
	"golang.org/x/crypto/acme/autocert"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Config is passed into NewSMCache as a way to configure how SMCache will behave
// through it's lifespan.
type Config struct {
	// ProjectID is the GCP Project ID where the Secrets will be stored.
	// This is the "Project ID" as seen in Google Cloud console.
	// Example ID: "my-project-1234".
	// This field is Required.
	ProjectID string

	// SecretPrefix is a string that will be put before the secret name.
	// This is useful for for IAM access control. As well, it's useful
	// for grouping secrets by application.
	// Optional, defaults to no-prefix.
	SecretPrefix string

	// If true, smcache will not delete old SecretVersions of Certificates.
	// If false, when autoert stores a certificate that is already in Secret Manager,
	// smcache will attempt to delete all old versions of that certificate.
	// Optional, defaults to false.
	KeepOldCertificates bool

	// DebugLogging controls if logging is enabled.
	// If true, smcache will log some status messages to log.Prtinf().
	// This will not logany sensitive data, it should just be key
	// names and paths.
	// Optional, defaults to false.
	DebugLogging bool
}

// smCache is the struct that will implement the autocert.Cache interface.
// It stores the needed data to interact with the GCP SecretManager.
type smCache struct {
	Config
	cf api.ClientFactory
}

// NewSMCache creates a struct that implements the `autocert.Cache` interface.
// It uses the Config passed in to drive the behavior of this client.
func NewSMCache(config Config) autocert.Cache {
	config.SecretPrefix = sanitize(config.SecretPrefix)

	return &smCache{
		Config: config,
		cf:     &api.SecretClientFactoryImpl{},
	}
}

// Get returns a certificate data for the specified key.
// If there's no such key, Get returns ErrCacheMiss.
func (smc *smCache) Get(ctx context.Context, key string) ([]byte, error) {
	key = sanitize(key)
	smc.logf("Get called for: [%v]", key)

	client, err := smc.cf.NewSecretClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to setup client: %w", err)
	}
	defer client.Close()

	svKey := fmt.Sprintf("projects/%s/secrets/%s%s/versions/latest", smc.ProjectID, smc.SecretPrefix, key)
	smc.logf("GET svKey: %v", svKey)

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: svKey,
	}
	resp, err := client.AccessSecretVersion(req)

	if st := status.Convert(err); st != nil {
		if st.Code() == codes.NotFound {
			return nil, autocert.ErrCacheMiss
		}

		return nil, err
	}

	smc.logf("GET: Got result: %+v", resp.GetName())

	return resp.GetPayload().GetData(), nil
}

// Only get the 10 most recent SecretVersions to delete for this secret.
const listPageSize = 10

// Put stores the data in the cache under the specified key.
// Underlying implementations may use any data storage format,
// as long as the reverse operation, Get, results in the original data.
func (smc *smCache) Put(ctx context.Context, key string, data []byte) error {
	key = sanitize(key)
	smc.logf("Put called for: [%v]", key)

	client, err := smc.cf.NewSecretClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to setup client: %w", err)
	}
	defer client.Close()

	// Get a List of SecretVersions that already exist in this secret.
	// If we get NotFound, we know to create the secret.
	// Otherwise we'll have a list of SecretVersions to delete once the rest is complete.
	svi := client.ListSecretVersions(&secretmanagerpb.ListSecretVersionsRequest{
		Parent: fmt.Sprintf("projects/%s/secrets/%s%s", smc.ProjectID, smc.SecretPrefix, key),
		// Should only need to get a few to delete.  Also hopefully they are returned in most-recent-first order
		PageSize: listPageSize,
	})

	sv, err := svi.Next()
	if st := status.Convert(err); st != nil {
		if st.Code() == codes.NotFound {
			// If the base Secret was NotFound, we attempt to create it
			err = smc.createSecret(key, client)
			if err != nil {
				// Secret creation failed, bail
				return err
			}
		} else {
			// Some other error happened, lets just return
			return err
		}
	}

	err = smc.addSecretVersion(key, data, client)
	if err != nil {
		return err
	}

	if !smc.KeepOldCertificates {
		smc.deleteOldSecretVersions(client, sv, svi)
	}

	return nil
}

// deleteOldSecretVersions will delete sv and all other SecretVersions within the svi.
// This is a best effort operation and will not return any errors if there are problems,
// but will log any problems (if debug logging is enabled).
func (smc *smCache) deleteOldSecretVersions(
	client api.SecretClient,
	sv *secretmanagerpb.SecretVersion,
	svi api.SecretListIterator) {
	var err error

	for {
		if sv == nil {
			return
		}

		// This code will only ever leave them in the "ENABLED" state,
		// so just try to delete those.
		if sv.GetState() == secretmanagerpb.SecretVersion_ENABLED {
			svr := &secretmanagerpb.DestroySecretVersionRequest{
				Name: sv.GetName(),
			}

			_, err = client.DestroySecretVersion(svr)
			if err != nil {
				smc.logf("Error deleting secret version: %v, got error %v", sv.GetName(), err)
			} else {
				smc.logf("Deleted secret %v", sv.GetName())
			}
		}
		// Get the next SecretVersion to delete
		sv, err = svi.Next()
		if err != nil {
			return
		}
	}
}

// createSecret will create the secret within the project.
func (smc *smCache) createSecret(key string, client api.SecretClient) error {
	createSecretReq := &secretmanagerpb.CreateSecretRequest{
		Parent:   fmt.Sprintf("projects/%s", smc.ProjectID),
		SecretId: fmt.Sprintf("%s%s", smc.SecretPrefix, key),
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	}

	_, err := client.CreateSecret(createSecretReq)
	if err != nil {
		return fmt.Errorf("failed to create Secret. %w", err)
	}

	return nil
}

// addSecretVersion will store the data within the secret
func (smc *smCache) addSecretVersion(key string, data []byte, client api.SecretClient) error {
	sKey := fmt.Sprintf("projects/%s/secrets/%s%s", smc.ProjectID, smc.SecretPrefix, key)

	req := &secretmanagerpb.AddSecretVersionRequest{
		Parent: sKey,
		Payload: &secretmanagerpb.SecretPayload{
			Data: data,
		},
	}

	_, err := client.AddSecretVersion(req)

	return err
}

// Delete removes a certificate data from the cache under the specified key.
// If there's no such key in the cache, Delete returns nil.
func (smc *smCache) Delete(ctx context.Context, key string) error {
	key = sanitize(key)
	smc.logf("Delete called for: [%v]", key)

	client, err := smc.cf.NewSecretClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to setup client: %w", err)
	}
	defer client.Close()

	sKey := fmt.Sprintf("projects/%s/secrets/%s%s", smc.ProjectID, smc.SecretPrefix, key)

	req := &secretmanagerpb.DeleteSecretRequest{
		Name: sKey,
	}

	err = client.DeleteSecret(req)
	if st := status.Convert(err); st != nil {
		// No-such-key, we return nil
		if st.Code() == codes.NotFound {
			return nil
		}
		// Some other problem happened while trying to delete, return the error
		return fmt.Errorf("problem while deleting secret [%v]. %w", sKey, err)
	}

	return nil
}

// logf to basic logger if DebugLogging is enabled.
func (smc *smCache) logf(format string, v ...interface{}) {
	if smc.DebugLogging {
		log.Printf("smcache: "+format, v...)
	}
}

// Secret Manager URLs are a bit picky about the characters that can be in them.
// This regex restricts the chars in the key passed in by autocert.
// NOTE: any changes to this regex will be a breaking change.
var allowedCharacters = regexp.MustCompile("[^a-zA-Z0-9-_]")

// Replace any non-URL safe characters with underscores.
func sanitize(s string) string {
	return allowedCharacters.ReplaceAllString(s, "_")
}
