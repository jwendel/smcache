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
	api "github.com/jwendel/smcache/internal"
	apimocks "github.com/jwendel/smcache/internal/mock"
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

	m := apimocks.NewMockSecretClient(ctrl)
	m.EXPECT().AccessSecretVersion(gomock.Any()).Return(nil, fmt.Errorf("Some random error"))
	m.EXPECT().Close().Times(1)

	cache := newCacheWithMockGrpc(Config{ProjectID: "a", SecretPrefix: "b", DebugLogging: debug}, m)
	data, err := cache.Get(context.Background(), "d")

	assert.EqualError(t, err, "Some random error")
	assert.Nil(t, data)
}

func TestGet_notFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := apimocks.NewMockSecretClient(ctrl)
	m.EXPECT().AccessSecretVersion(gomock.Any()).Return(nil, status.Error(codes.NotFound, "fake not found"))
	m.EXPECT().Close().Times(1)

	cache := newCacheWithMockGrpc(Config{ProjectID: "a", SecretPrefix: "b", DebugLogging: debug}, m)
	data, err := cache.Get(context.Background(), "d")

	assert.EqualError(t, err, autocert.ErrCacheMiss.Error())
	assert.Nil(t, data)
}

func TestGet_happyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	secret := []byte("secret data!")
	m := apimocks.NewMockSecretClient(ctrl)
	m.EXPECT().AccessSecretVersion(gomock.Eq(
		&secretmanagerpb.AccessSecretVersionRequest{
			Name: "projects/a/secrets/bd/versions/latest",
		})).Return(
		&secretmanagerpb.AccessSecretVersionResponse{
			Name:    "bd",
			Payload: &secretmanagerpb.SecretPayload{Data: secret},
		}, nil)
	m.EXPECT().Close().Times(1)

	cache := newCacheWithMockGrpc(Config{ProjectID: "a", SecretPrefix: "b", DebugLogging: debug}, m)
	result, err := cache.Get(context.Background(), "d")

	assert.Nil(t, err)
	assert.Equal(t, result, secret)
}

func TestGet_happyPath_sanatizeKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	secret := []byte("secret data!")
	m := apimocks.NewMockSecretClient(ctrl)
	m.EXPECT().AccessSecretVersion(gomock.Eq(
		&secretmanagerpb.AccessSecretVersionRequest{
			Name: "projects/a!@#$_^&*()-/secrets/b_________-_d___________/versions/latest",
		})).Return(
		&secretmanagerpb.AccessSecretVersionResponse{
			Name:    "bd",
			Payload: &secretmanagerpb.SecretPayload{Data: secret},
		}, nil)
	m.EXPECT().Close().Times(1)

	cache := newCacheWithMockGrpc(Config{ProjectID: "a!@#$_^&*()-", SecretPrefix: `b.)(*&^$#@-_`, DebugLogging: debug}, m) //
	result, err := cache.Get(context.Background(), "d!@#$$%^&*()")

	assert.Nil(t, err)
	assert.Equal(t, result, secret)
}

func TestGet_unsetPrefix(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	secret := []byte("secret data!")
	m := apimocks.NewMockSecretClient(ctrl)
	m.EXPECT().AccessSecretVersion(gomock.Eq(
		&secretmanagerpb.AccessSecretVersionRequest{
			Name: "projects/a/secrets/d/versions/latest",
		})).Return(
		&secretmanagerpb.AccessSecretVersionResponse{
			Name:    "d",
			Payload: &secretmanagerpb.SecretPayload{Data: secret},
		}, nil)
	m.EXPECT().Close().Times(1)

	cache := newCacheWithMockGrpc(Config{ProjectID: "a", DebugLogging: debug}, m)
	result, err := cache.Get(context.Background(), "d")

	assert.Nil(t, err)
	assert.Equal(t, result, secret)
}

func TestGet_clientError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := apimocks.NewMockSecretClient(ctrl)

	cache := newCacheWithErrorMockGrpc(Config{ProjectID: "a", DebugLogging: debug}, m)
	result, err := cache.Get(context.Background(), "d")

	assert.EqualError(t, err, "failed to setup client: problem creating client")
	assert.Nil(t, result)
}

func TestPut_happyPath_noSecretVersions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	secret := []byte("secret data!")
	secretPath := "projects/a/secrets/d"
	m := apimocks.NewMockSecretClient(ctrl)
	m.EXPECT().ListSecretVersions(gomock.Eq(
		&secretmanagerpb.ListSecretVersionsRequest{
			Parent:   secretPath,
			PageSize: 10,
		})).Return(
		&sliFake{})
	m.EXPECT().AddSecretVersion(gomock.Eq(&secretmanagerpb.AddSecretVersionRequest{
		Parent:  secretPath,
		Payload: &secretmanagerpb.SecretPayload{Data: secret},
	})).Return(nil, nil)
	m.EXPECT().Close().Times(1)

	cache := newCacheWithMockGrpc(Config{ProjectID: "a", DebugLogging: debug}, m)
	err := cache.Put(context.Background(), "d", secret)

	assert.Nil(t, err)
}

func TestPut_happyPath_oneSV(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	secret := []byte("secret data!")
	secretPath := "projects/a/secrets/d"
	activeSV := secretPath + "/versions/4"
	m := apimocks.NewMockSecretClient(ctrl)
	m.EXPECT().ListSecretVersions(gomock.Eq(
		&secretmanagerpb.ListSecretVersionsRequest{
			Parent:   secretPath,
			PageSize: 10,
		})).Return(
		&sliFake{secrets: []*secretmanagerpb.SecretVersion{{Name: activeSV, State: secretmanagerpb.SecretVersion_ENABLED}}})
	m.EXPECT().AddSecretVersion(gomock.Eq(&secretmanagerpb.AddSecretVersionRequest{
		Parent:  secretPath,
		Payload: &secretmanagerpb.SecretPayload{Data: secret},
	})).Return(nil, nil)
	m.EXPECT().DestroySecretVersion(gomock.Eq(&secretmanagerpb.DestroySecretVersionRequest{
		Name: activeSV,
	})).Return(nil, nil)
	m.EXPECT().Close().Times(1)

	cache := newCacheWithMockGrpc(Config{ProjectID: "a", DebugLogging: debug}, m)
	err := cache.Put(context.Background(), "d", secret)

	assert.Nil(t, err)
}

// helper functions for tests

// Iterator fakes

type sliFake struct {
	secrets []*secretmanagerpb.SecretVersion
}

func (sl *sliFake) Next() (*secretmanagerpb.SecretVersion, error) {
	if len(sl.secrets) > 0 {
		s := sl.secrets[0]
		sl.secrets[0] = nil
		sl.secrets = sl.secrets[1:]
		return s, nil
	}

	return nil, nil
}

// GRPC mocks

func newCacheWithMockGrpc(config Config, m *apimocks.MockSecretClient) *smCache {
	c := NewSMCache(config).(*smCache)
	c.cf = &mockSecretClientFactoryImpl{mock: m}
	return c
}

type mockSecretClientFactoryImpl struct {
	mock *apimocks.MockSecretClient
}

func (m *mockSecretClientFactoryImpl) NewSecretClient(ctx context.Context) (api.SecretClient, error) {
	return m.mock, nil
}

func newCacheWithErrorMockGrpc(config Config, m *apimocks.MockSecretClient) *smCache {
	c := NewSMCache(config).(*smCache)
	c.cf = &mockErrorSecretClientFactoryImpl{mock: m}
	return c
}

type mockErrorSecretClientFactoryImpl struct {
	mock *apimocks.MockSecretClient
}

func (m *mockErrorSecretClientFactoryImpl) NewSecretClient(ctx context.Context) (api.SecretClient, error) {
	return nil, fmt.Errorf("problem creating client")
}
