package gsmcache

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jwendel/gsmcache/mocks"
)

func TestGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cache := newGSMCacheWithMock(GSMCacheConfig{ProjectID: "a", SecretPrefix: "b", DebugLogging: false}, ctrl)
	data, err := cache.Get(context.Background(), "d")
	if err != nil {
		t.Error(err)
	} else {
		t.Log(data)
	}

}

func newGSMCacheWithMock(config GSMCacheConfig, ctrl *gomock.Controller) *GSMCache {
	return &GSMCache{
		GSMCacheConfig: config,
		cf:             &mockSecretClientFactoryImpl{ctrl: ctrl},
	}
}

type mockSecretClientFactoryImpl struct {
	ctrl *gomock.Controller
}

func (m *mockSecretClientFactoryImpl) newSecretClient(ctx context.Context) (secretClient, error) {
	return mocks.NewMocksecretClient(m.ctrl), nil
}
