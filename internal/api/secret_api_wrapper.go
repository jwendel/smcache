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

	sm "cloud.google.com/go/secretmanager/apiv1"
	smpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

// ClientFactory is used to create SecretClient, which is the GRPC Secret Client
// in normal use, but can be mocked for tests.
type ClientFactory interface {
	NewSecretClient(ctx context.Context) (SecretClient, error)
}

// SecretClient is a wrapper around the secretmanager APIs that are used by smcache.
// It is entirely for the purpose of being able to mock these for testing.
type SecretClient interface {
	AccessSecretVersion(req *smpb.AccessSecretVersionRequest) (*smpb.AccessSecretVersionResponse, error)
	ListSecretVersions(req *smpb.ListSecretVersionsRequest) SecretListIterator
	DestroySecretVersion(req *smpb.DestroySecretVersionRequest) (*smpb.SecretVersion, error)
	CreateSecret(req *smpb.CreateSecretRequest) (*smpb.Secret, error)
	AddSecretVersion(req *smpb.AddSecretVersionRequest) (*smpb.SecretVersion, error)
	DeleteSecret(req *smpb.DeleteSecretRequest) error
	Close() error
}

// SecretListIterator is an interface for the GRPC secret manager response from ListSecretVersions.
type SecretListIterator interface {
	Next() (*smpb.SecretVersion, error)
}

type secretClientImpl struct {
	client *sm.Client
	ctx    context.Context
}

// SecretClientFactoryImpl implements ClientFactory for the real GRPC client.
type SecretClientFactoryImpl struct{}

// NewSecretClient creates a GRPC NewClient for secretmanager.
func (*SecretClientFactoryImpl) NewSecretClient(ctx context.Context) (SecretClient, error) {
	c, err := sm.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to setup client: %w", err)
	}

	return &secretClientImpl{client: c, ctx: ctx}, nil
}

func (sc *secretClientImpl) AccessSecretVersion(req *smpb.AccessSecretVersionRequest) (*smpb.AccessSecretVersionResponse, error) {
	return sc.client.AccessSecretVersion(sc.ctx, req)
}
func (sc *secretClientImpl) ListSecretVersions(req *smpb.ListSecretVersionsRequest) SecretListIterator {
	return sc.client.ListSecretVersions(sc.ctx, req)
}
func (sc *secretClientImpl) DestroySecretVersion(req *smpb.DestroySecretVersionRequest) (*smpb.SecretVersion, error) {
	return sc.client.DestroySecretVersion(sc.ctx, req)
}
func (sc *secretClientImpl) CreateSecret(req *smpb.CreateSecretRequest) (*smpb.Secret, error) {
	return sc.client.CreateSecret(sc.ctx, req)
}
func (sc *secretClientImpl) AddSecretVersion(req *smpb.AddSecretVersionRequest) (*smpb.SecretVersion, error) {
	return sc.client.AddSecretVersion(sc.ctx, req)
}
func (sc *secretClientImpl) DeleteSecret(req *smpb.DeleteSecretRequest) error {
	return sc.client.DeleteSecret(sc.ctx, req)
}

func (sc *secretClientImpl) Close() error {
	return sc.client.Close()
}
