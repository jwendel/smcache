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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/smcache/mocks"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/acme/autocert"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const debug = false

func TestGet_errResp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMocksecretClient(ctrl)
	m.EXPECT().AccessSecretVersion(gomock.Any()).Return(nil, fmt.Errorf("Some random error"))

	cache := newCacheWithMockGrpc(Config{ProjectID: "a", SecretPrefix: "b", DebugLogging: debug}, m)
	data, err := cache.Get(context.Background(), "d")

	assert.EqualError(t, err, "Some random error")
	assert.Nil(t, data)
}

func TestGet_notFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMocksecretClient(ctrl)
	m.EXPECT().AccessSecretVersion(gomock.Any()).Return(nil, status.Error(codes.NotFound, "fake not found"))

	cache := newCacheWithMockGrpc(Config{ProjectID: "a", SecretPrefix: "b", DebugLogging: debug}, m)
	data, err := cache.Get(context.Background(), "d")

	assert.EqualError(t, err, autocert.ErrCacheMiss.Error())
	assert.Nil(t, data)
}

func TestGet_happyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	secret := []byte("secret data!")
	m := mocks.NewMocksecretClient(ctrl)
	m.EXPECT().AccessSecretVersion(gomock.Eq(
		&secretmanagerpb.AccessSecretVersionRequest{
			Name: "projects/a/secrets/bd/versions/latest",
		})).Return(
		&secretmanagerpb.AccessSecretVersionResponse{
			Name:    "bd",
			Payload: &secretmanagerpb.SecretPayload{Data: secret},
		}, nil)

	cache := newCacheWithMockGrpc(Config{ProjectID: "a", SecretPrefix: "b", DebugLogging: debug}, m)
	result, err := cache.Get(context.Background(), "d")

	assert.Nil(t, err)
	assert.Equal(t, result, secret)
}

// helper functions for tests

func newCacheWithMockGrpc(config Config, m *mocks.MocksecretClient) *smCache {
	return &smCache{
		Config: config,
		cf:     &mockSecretClientFactoryImpl{mock: m},
	}
}

type mockSecretClientFactoryImpl struct {
	mock *mocks.MocksecretClient
}

func (m *mockSecretClientFactoryImpl) newSecretClient(ctx context.Context) (secretClient, error) {
	return m.mock, nil
}
