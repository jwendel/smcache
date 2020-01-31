package gsmcache

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1beta1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1beta1"
)

type clientFactory interface {
	newSecretClient(ctx context.Context) (secretClient, error)
}

// secretClient is a wrapper around the secretmanager APIs that are used by gsmcache.
// It is entirely for the purpose of being able to mock these for testing.
type secretClient interface {
	AccessSecretVersion(req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error)
	ListSecretVersions(req *secretmanagerpb.ListSecretVersionsRequest) *secretmanager.SecretVersionIterator
	DestroySecretVersion(req *secretmanagerpb.DestroySecretVersionRequest) (*secretmanagerpb.SecretVersion, error)
	CreateSecret(req *secretmanagerpb.CreateSecretRequest) (*secretmanagerpb.Secret, error)
	AddSecretVersion(req *secretmanagerpb.AddSecretVersionRequest) (*secretmanagerpb.SecretVersion, error)
	DeleteSecret(req *secretmanagerpb.DeleteSecretRequest) error
}

type secretClientImpl struct {
	client *secretmanager.Client
	ctx    context.Context
}

type secretClientFactoryImpl struct{}

func (*secretClientFactoryImpl) newSecretClient(ctx context.Context) (secretClient, error) {
	c, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to setup client: %w", err)
	}

	return &secretClientImpl{client: c, ctx: ctx}, nil
}

func (sc *secretClientImpl) AccessSecretVersion(req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	return sc.client.AccessSecretVersion(sc.ctx, req)
}
func (sc *secretClientImpl) ListSecretVersions(req *secretmanagerpb.ListSecretVersionsRequest) *secretmanager.SecretVersionIterator {
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
