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

	secretmanager "cloud.google.com/go/secretmanager/apiv1beta1"
	"golang.org/x/crypto/acme/autocert"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Config represents a autocert cache that will store/retreive data from
// GCP Secret Manager.  TODO, reword me
type Config struct {
	// ProjectID is the GCP Project ID where the Secrets will be stored
	ProjectID string
	// SecretPrefix is a string that will be put before the secret name.
	// This is useful for for IAM access control and for grouping secrets
	// by application.
	SecretPrefix string
	// If true, will log some status messages to log.Prtinf().
	DebugLogging bool
}

// smCache should be private.  TODO
type smCache struct {
	Config
	cf clientFactory
}

// NewSMCache creates a new smcache TODO
// TODO: Change this to return an interface
func NewSMCache(config Config) autocert.Cache {
	return &smCache{
		Config: config,
		cf:     &secretClientFactoryImpl{},
	}
}

// Get returns a certificate data for the specified key.
// If there's no such key, Get returns ErrCacheMiss.
func (smc *smcache) Get(ctx context.Context, key string) ([]byte, error) {
	key = sanitize(key)

	smc.dlog("Get called for: [%v]", key)
	client, err := smc.cf.newSecretClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to setup client: %w", err)
	}

	svKey := fmt.Sprintf("projects/%s/secrets/%s%s/versions/latest", smc.ProjectID, smc.SecretPrefix, key)
	smc.dlog("GET svKey: %v", svKey)

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

	smc.dlog("GET: Got result: %+v", resp.GetName())
	return resp.GetPayload().GetData(), nil
}

// Put stores the data in the cache under the specified key.
// Underlying implementations may use any data storage format,
// as long as the reverse operation, Get, results in the original data.
func (smc *smcache) Put(ctx context.Context, key string, data []byte) error {
	key = sanitize(key)

	smc.dlog("Put called for: [%v]", key)
	client, err := smc.cf.newSecretClient(ctx)
	if err != nil {
		return fmt.Errorf("Failed to setup client: %w", err)
	}

	// Get a List of SecretVersions that already exist in this secret.
	// If we get NotFound, we know to create the secret.
	// Otherwise we'll have a list of SecretVersions to delete once the rest is complete.
	svi := client.ListSecretVersions(&secretmanagerpb.ListSecretVersionsRequest{
		Parent: fmt.Sprintf("projects/%s/secrets/%s%s", smc.ProjectID, smc.SecretPrefix, key),
		// Should only need to get a few to delete.  Also hopefully they are returned in most-recent-first order
		PageSize: 10,
	})

	sv, err := svi.Next()
	if st := status.Convert(err); st != nil {
		if st.Code() == codes.NotFound {
			// If the base Secret was NotFound, we attempt to create it
			err := smc.createSecret(key, client)
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

	smc.deleteOldSecretVersions(key, client, sv, svi)
	return nil
}

// deleteOldSecretVersions will delete sv and all other SecretVersions within the svi.
// This is a best effort operation and will not return any errors if there are problems,
// but will log any problems (if debug logging is enabled).
func (smc *smcache) deleteOldSecretVersions(
	key string,
	client secretClient,
	sv *secretmanagerpb.SecretVersion,
	svi *secretmanager.SecretVersionIterator) {

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
			_, err := client.DestroySecretVersion(svr)
			if err != nil {
				smc.dlog("error deleting secret version: %v, got error %v", sv.GetName(), err)
				return
			}
			smc.dlog("Deleted secret %v", sv.GetName())
		}
		// Get the next SecretVersion to delete
		var err error
		sv, err = svi.Next()
		if err != nil {
			return
		}
	}
}

// createSecret will create the secret within the project.
func (smc *smcache) createSecret(key string, client secretClient) error {
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
		return fmt.Errorf("Failed to create Secret. %w", err)
	}
	return nil
}

// addSecretVersion will store the data within the secret
func (smc *smcache) addSecretVersion(key string, data []byte, client secretClient) error {
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
func (smc *smcache) Delete(ctx context.Context, key string) error {
	key = sanitize(key)

	smc.dlog("Delete called for: [%v]", key)
	client, err := smc.cf.newSecretClient(ctx)
	if err != nil {
		return fmt.Errorf("Failed to setup client: %w", err)
	}

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
		return fmt.Errorf("Problem while deleting secret [%v]. %w", sKey, err)
	}

	return nil
}

func (smc *smcache) dlog(format string, v ...interface{}) {
	if smc.DebugLogging {
		log.Printf(format, v...)
	}
}

// Secret Manager URLs are a bit picky about the characters that can be in them.
// This regex restricts the chars in the key passed in by autocert.
// NOTE: any changes to this regex will be a breaking change.
var allowedCharacters = regexp.MustCompile("[^a-zA-Z0-9-_]+")

// Replace any non-URL safe characters with underscores.
func sanitize(s string) string {
	return allowedCharacters.ReplaceAllString(s, "_")
}
