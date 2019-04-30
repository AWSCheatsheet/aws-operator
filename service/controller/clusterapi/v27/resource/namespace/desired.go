package namespace

import (
	"context"

	"github.com/giantswarm/microerror"
	apiv1 "k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/legacykey"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := legacykey.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Compute the desired state of the namespace to have a reference of how
	// the data should be.
	namespace := &apiv1.Namespace{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name: legacykey.ClusterNamespace(customObject),
			Labels: map[string]string{
				"cluster":  legacykey.ClusterID(customObject),
				"customer": legacykey.OrganizationID(customObject),
			},
		},
	}

	return namespace, nil
}
