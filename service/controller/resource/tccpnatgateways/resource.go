package tccpnatgateways

import (
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "tccpnatgateways"
)

type Config struct {
	Logger        micrologger.Logger
	ToClusterFunc func(ctx context.Context, v interface{}) (infrastructurev1alpha2.AWSCluster, error)
}

type Resource struct {
	logger        micrologger.Logger
	toClusterFunc func(ctx context.Context, v interface{}) (infrastructurev1alpha2.AWSCluster, error)
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.ToClusterFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterFunc must not be empty", config)
	}

	r := &Resource{
		logger:        config.Logger,
		toClusterFunc: config.ToClusterFunc,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
