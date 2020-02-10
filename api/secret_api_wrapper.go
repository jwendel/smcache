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

package api

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1beta1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1beta1"
)

type ClientFactory interface {
	NewSecretClient(ctx context.Context) (SecretClient, error)
}

// SecretClient is a wrapper around the secretmanager APIs that are used by smcache.
// It is entirely for the purpose of being able to mock these for testing.
type SecretClient interface {
	AccessSecretVersion(req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error)
	ListSecretVersions(req *secretmanagerpb.ListSecretVersionsRequest) SecretListIterator
	DestroySecretVersion(req *secretmanagerpb.DestroySecretVersionRequest) (*secretmanagerpb.SecretVersion, error)
	CreateSecret(req *secretmanagerpb.CreateSecretRequest) (*secretmanagerpb.Secret, error)
	AddSecretVersion(req *secretmanagerpb.AddSecretVersionRequest) (*secretmanagerpb.SecretVersion, error)
	DeleteSecret(req *secretmanagerpb.DeleteSecretRequest) error
}

type SecretListIterator interface {
	Next() (*secretmanagerpb.SecretVersion, error)
}

type secretClientImpl struct {
	client *secretmanager.Client
	ctx    context.Context
}

type SecretClientFactoryImpl struct{}

func (*SecretClientFactoryImpl) NewSecretClient(ctx context.Context) (SecretClient, error) {
	c, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to setup client: %w", err)
	}

	return &secretClientImpl{client: c, ctx: ctx}, nil
}

func (sc *secretClientImpl) AccessSecretVersion(req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	return sc.client.AccessSecretVersion(sc.ctx, req)
}
func (sc *secretClientImpl) ListSecretVersions(req *secretmanagerpb.ListSecretVersionsRequest) SecretListIterator {
	return sc.client.ListSecretVersions(sc.ctx, req)
}
func (sc *secretClientImpl) DestroySecretVersion(req *secretmanagerpb.DestroySecretVersionRequest) (*secretmanagerpb.SecretVersion, error) {
	return sc.client.DestroySecretVersion(sc.ctx, req)
}
func (sc *secretClientImpl) CreateSecret(req *secretmanagerpb.CreateSecretRequest) (*secretmanagerpb.Secret, error) {
	return sc.client.CreateSecret(sc.ctx, req)
}
func (sc *secretClientImpl) AddSecretVersion(req *secretmanagerpb.AddSecretVersionRequest) (*secretmanagerpb.SecretVersion, error) {
	return sc.client.AddSecretVersion(sc.ctx, req)
}
func (sc *secretClientImpl) DeleteSecret(req *secretmanagerpb.DeleteSecretRequest) error {
	return sc.client.DeleteSecret(sc.ctx, req)
}
