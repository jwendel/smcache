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
	"github.com/jwendel/smcache/internal/api"
	apimocks "github.com/jwendel/smcache/internal/api/mock"
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

// Make sure the looping delete is tested (deleteOldSecretVersions)
func TestPut_happyPath_5_SVs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	secret := []byte("secret data!")
	secretPath := "projects/a/secrets/d"
	activeSV := secretPath + "/versions/"
	m := apimocks.NewMockSecretClient(ctrl)
	m.EXPECT().ListSecretVersions(gomock.Eq(
		&secretmanagerpb.ListSecretVersionsRequest{
			Parent:   secretPath,
			PageSize: 10,
		})).Return(
		&sliFake{secrets: []*secretmanagerpb.SecretVersion{
			{Name: activeSV + "1", State: secretmanagerpb.SecretVersion_ENABLED},
			{Name: activeSV + "2", State: secretmanagerpb.SecretVersion_DESTROYED},
			{Name: activeSV + "3", State: secretmanagerpb.SecretVersion_ENABLED},
			{Name: activeSV + "4", State: secretmanagerpb.SecretVersion_DISABLED},
			{Name: activeSV + "5", State: secretmanagerpb.SecretVersion_ENABLED},
		}})
	m.EXPECT().AddSecretVersion(gomock.Eq(&secretmanagerpb.AddSecretVersionRequest{
		Parent:  secretPath,
		Payload: &secretmanagerpb.SecretPayload{Data: secret},
	})).Return(nil, nil)
	m.EXPECT().DestroySecretVersion(gomock.Eq(&secretmanagerpb.DestroySecretVersionRequest{
		Name: activeSV + "1",
	})).Return(nil, nil)
	m.EXPECT().DestroySecretVersion(gomock.Eq(&secretmanagerpb.DestroySecretVersionRequest{
		Name: activeSV + "3",
	})).Return(nil, fmt.Errorf("Fake error"))
	m.EXPECT().DestroySecretVersion(gomock.Eq(&secretmanagerpb.DestroySecretVersionRequest{
		Name: activeSV + "5",
	})).Return(nil, nil)
	m.EXPECT().Close().Times(1)

	cache := newCacheWithMockGrpc(Config{ProjectID: "a", DebugLogging: debug}, m)
	err := cache.Put(context.Background(), "d", secret)

	assert.Nil(t, err)
}

// Make sure not to delete old ones if it's a list of
// multiple when KeepOldCertificates is enabled.
func TestPut_happyPath_5_SVs_keepSVs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	secret := []byte("secret data!")
	secretPath := "projects/a/secrets/d"
	activeSV := secretPath + "/versions/"
	m := apimocks.NewMockSecretClient(ctrl)
	m.EXPECT().ListSecretVersions(gomock.Eq(
		&secretmanagerpb.ListSecretVersionsRequest{
			Parent:   secretPath,
			PageSize: 10,
		})).Return(
		&sliFake{secrets: []*secretmanagerpb.SecretVersion{
			{Name: activeSV + "1", State: secretmanagerpb.SecretVersion_ENABLED},
			{Name: activeSV + "2", State: secretmanagerpb.SecretVersion_DESTROYED},
			{Name: activeSV + "3", State: secretmanagerpb.SecretVersion_ENABLED},
			{Name: activeSV + "4", State: secretmanagerpb.SecretVersion_DISABLED},
			{Name: activeSV + "5", State: secretmanagerpb.SecretVersion_ENABLED},
		}})
	m.EXPECT().AddSecretVersion(gomock.Eq(&secretmanagerpb.AddSecretVersionRequest{
		Parent:  secretPath,
		Payload: &secretmanagerpb.SecretPayload{Data: secret},
	})).Return(nil, nil)
	m.EXPECT().Close().Times(1)

	cache := newCacheWithMockGrpc(Config{ProjectID: "a", KeepOldCertificates: true, DebugLogging: debug}, m)
	err := cache.Put(context.Background(), "d", secret)

	assert.Nil(t, err)
}

func TestPut_happyPath_NewSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	secret := []byte("secret data!")
	secretPath := "projects/projId/secrets/secrId"
	m := apimocks.NewMockSecretClient(ctrl)
	m.EXPECT().ListSecretVersions(gomock.Eq(
		&secretmanagerpb.ListSecretVersionsRequest{
			Parent:   secretPath,
			PageSize: 10,
		})).Return(
		&sliFakeNotFound{})
	m.EXPECT().CreateSecret(&secretmanagerpb.CreateSecretRequest{
		Parent:   "projects/projId",
		SecretId: "secrId",
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	}).Return(nil, nil)
	m.EXPECT().AddSecretVersion(gomock.Eq(&secretmanagerpb.AddSecretVersionRequest{
		Parent:  secretPath,
		Payload: &secretmanagerpb.SecretPayload{Data: secret},
	})).Return(nil, nil)
	m.EXPECT().Close().Times(1)

	cache := newCacheWithMockGrpc(Config{ProjectID: "projId", DebugLogging: debug}, m)
	err := cache.Put(context.Background(), "secrId", secret)

	assert.Nil(t, err)
}

func TestPut_NewSecret_createError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	secret := []byte("secret data!")
	secretPath := "projects/projId/secrets/secrId"
	m := apimocks.NewMockSecretClient(ctrl)
	m.EXPECT().ListSecretVersions(gomock.Eq(
		&secretmanagerpb.ListSecretVersionsRequest{
			Parent:   secretPath,
			PageSize: 10,
		})).Return(
		&sliFakeNotFound{})
	m.EXPECT().CreateSecret(&secretmanagerpb.CreateSecretRequest{
		Parent:   "projects/projId",
		SecretId: "secrId",
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	}).Return(nil, fmt.Errorf("create error"))
	m.EXPECT().Close().Times(1)

	cache := newCacheWithMockGrpc(Config{ProjectID: "projId", DebugLogging: debug}, m)
	err := cache.Put(context.Background(), "secrId", secret)

	assert.EqualError(t, err, "failed to create Secret. create error")
}

func TestPut_listError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	secret := []byte("secret data!")
	secretPath := "projects/projId/secrets/secrId"
	m := apimocks.NewMockSecretClient(ctrl)
	m.EXPECT().ListSecretVersions(gomock.Eq(
		&secretmanagerpb.ListSecretVersionsRequest{
			Parent:   secretPath,
			PageSize: 10,
		})).Return(
		&sliFakeError{})
	m.EXPECT().Close().Times(1)

	cache := newCacheWithMockGrpc(Config{ProjectID: "projId", DebugLogging: debug}, m)
	err := cache.Put(context.Background(), "secrId", secret)

	assert.EqualError(t, err, "rpc error: code = DeadlineExceeded desc = deadline exceeded resp")
}

func TestPut_errorOnSecretVersionCreate(t *testing.T) {
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
	})).Return(nil, fmt.Errorf("sv create error"))
	m.EXPECT().Close().Times(1)

	cache := newCacheWithMockGrpc(Config{ProjectID: "a", DebugLogging: debug}, m)
	err := cache.Put(context.Background(), "d", secret)

	assert.EqualError(t, err, "sv create error")
}

func TestPut_happyPath_KeepOldCertificates(t *testing.T) {
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
	m.EXPECT().Close().Times(1)

	cache := newCacheWithMockGrpc(Config{ProjectID: "a", DebugLogging: debug, KeepOldCertificates: true}, m)
	err := cache.Put(context.Background(), "d", secret)

	assert.Nil(t, err)
}

func TestPut_clientError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := apimocks.NewMockSecretClient(ctrl)

	secret := []byte("secret data!")
	cache := newCacheWithErrorMockGrpc(Config{ProjectID: "a", DebugLogging: debug}, m)
	err := cache.Put(context.Background(), "keykey", secret)

	assert.EqualError(t, err, "failed to setup client: problem creating client")
}

func TestDelete_clientError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := apimocks.NewMockSecretClient(ctrl)

	cache := newCacheWithErrorMockGrpc(Config{ProjectID: "a", DebugLogging: debug}, m)
	err := cache.Delete(context.Background(), "Key")

	assert.EqualError(t, err, "failed to setup client: problem creating client")
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

type sliFakeNotFound struct{}

func (sl *sliFakeNotFound) Next() (*secretmanagerpb.SecretVersion, error) {
	return nil, status.Error(codes.NotFound, "not found resp")
}

type sliFakeError struct{}

func (sl *sliFakeError) Next() (*secretmanagerpb.SecretVersion, error) {
	return nil, status.Error(codes.DeadlineExceeded, "deadline exceeded resp")
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
