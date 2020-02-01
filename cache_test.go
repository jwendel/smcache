package gsmcache

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jwendel/gsmcache/mocks"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/acme/autocert"
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

func newCacheWithMockGrpc(config Config, m *mocks.MocksecretClient) *gsmCache {
	return &gsmCache{
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
