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

	secretmanager "cloud.google.com/go/secretmanager/apiv1beta1"
	"github.com/golang/mock/gomock"
	"github.com/jwendel/smcache/mocks"
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

func TestGet_happyPath_sanatizeKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	secret := []byte("secret data!")
	m := mocks.NewMocksecretClient(ctrl)
	m.EXPECT().AccessSecretVersion(gomock.Eq(
		&secretmanagerpb.AccessSecretVersionRequest{
			Name: "projects/a!@#$_^&*()-/secrets/b_________-_d___________/versions/latest",
		})).Return(
		&secretmanagerpb.AccessSecretVersionResponse{
			Name:    "bd",
			Payload: &secretmanagerpb.SecretPayload{Data: secret},
		}, nil)

	cache := newCacheWithMockGrpc(Config{ProjectID: "a!@#$_^&*()-", SecretPrefix: `b.)(*&^$#@-_`, DebugLogging: debug}, m) //
	result, err := cache.Get(context.Background(), "d!@#$$%^&*()")

	assert.Nil(t, err)
	assert.Equal(t, result, secret)
}

func TestGet_unsetPrefix(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	secret := []byte("secret data!")
	m := mocks.NewMocksecretClient(ctrl)
	m.EXPECT().AccessSecretVersion(gomock.Eq(
		&secretmanagerpb.AccessSecretVersionRequest{
			Name: "projects/a/secrets/d/versions/latest",
		})).Return(
		&secretmanagerpb.AccessSecretVersionResponse{
			Name:    "d",
			Payload: &secretmanagerpb.SecretPayload{Data: secret},
		}, nil)

	cache := newCacheWithMockGrpc(Config{ProjectID: "a", DebugLogging: debug}, m)
	result, err := cache.Get(context.Background(), "d")

	assert.Nil(t, err)
	assert.Equal(t, result, secret)
}

func TestGet_clientError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMocksecretClient(ctrl)

	cache := newCacheWithErrorMockGrpc(Config{ProjectID: "a", DebugLogging: debug}, m)
	result, err := cache.Get(context.Background(), "d")

	assert.EqualError(t, err, "Failed to setup client: problem creating client")
	assert.Nil(t, result)
}

func TestPut(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	secret := []byte("secret data!")
	m := mocks.NewMocksecretClient(ctrl)
	m.EXPECT().ListSecretVersions(gomock.Eq(
		&secretmanagerpb.ListSecretVersionsRequest{
			Parent: "",
		})).Return(
		&secretmanager.SecretVersionIterator{}) // HOW DO I MOCK U?

	cache := newCacheWithMockGrpc(Config{ProjectID: "a", DebugLogging: debug}, m)
	err := cache.Put(context.Background(), "d", secret)

	assert.Nil(t, err)
}

// helper functions for tests

func newCacheWithMockGrpc(config Config, m *mocks.MocksecretClient) *smCache {
	c := NewSMCache(config).(*smCache)
	c.cf = &mockSecretClientFactoryImpl{mock: m}
	return c
}

type mockSecretClientFactoryImpl struct {
	mock *mocks.MocksecretClient
}

func (m *mockSecretClientFactoryImpl) newSecretClient(ctx context.Context) (secretClient, error) {
	return m.mock, nil
}

func newCacheWithErrorMockGrpc(config Config, m *mocks.MocksecretClient) *smCache {
	c := NewSMCache(config).(*smCache)
	c.cf = &mockErrorSecretClientFactoryImpl{mock: m}
	return c
}

type mockErrorSecretClientFactoryImpl struct {
	mock *mocks.MocksecretClient
}

func (m *mockErrorSecretClientFactoryImpl) newSecretClient(ctx context.Context) (secretClient, error) {
	return nil, fmt.Errorf("problem creating client")
}
