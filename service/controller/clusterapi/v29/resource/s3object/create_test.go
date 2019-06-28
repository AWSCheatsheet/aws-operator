package s3object

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/certs/certstest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/randomkeys/randomkeystest"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/controllercontext"
)

func Test_Resource_S3Object_newCreate(t *testing.T) {
	testCases := []struct {
		description   string
		obj           interface{}
		currentState  map[string]BucketObjectState
		desiredState  map[string]BucketObjectState
		expectedState map[string]BucketObjectState
	}{
		{
			description: "current state empty, desired state empty, empty create change",
			obj: &v1alpha1.Cluster{
				Status: v1alpha1.ClusterStatus{
					ProviderStatus: &runtime.RawExtension{
						Raw: []byte(`
							{
								"cluster": {
									"id": "5xchu"
								}
							}
						`),
					},
				},
			},
			currentState:  map[string]BucketObjectState{},
			desiredState:  map[string]BucketObjectState{},
			expectedState: map[string]BucketObjectState{},
		},
		{
			description: "current state empty, desired state not empty, create change == desired state",
			obj: &v1alpha1.Cluster{
				Status: v1alpha1.ClusterStatus{
					ProviderStatus: &runtime.RawExtension{
						Raw: []byte(`
							{
								"cluster": {
									"id": "5xchu"
								}
							}
						`),
					},
				},
			},
			currentState: map[string]BucketObjectState{},
			desiredState: map[string]BucketObjectState{
				"master": {
					Body:   "master-body",
					Bucket: "mybucket",
					Key:    "master",
				},
			},
			expectedState: map[string]BucketObjectState{
				"master": {
					Body:   "master-body",
					Bucket: "mybucket",
					Key:    "master",
				},
			},
		},
		{
			description: "current state not empty, desired state not empty, create change == desired state",
			obj: &v1alpha1.Cluster{
				Status: v1alpha1.ClusterStatus{
					ProviderStatus: &runtime.RawExtension{
						Raw: []byte(`
							{
								"cluster": {
									"id": "5xchu"
								}
							}
						`),
					},
				},
			},
			currentState: map[string]BucketObjectState{
				"mykey": {
					Body:   "mykey",
					Bucket: "mybucket",
					Key:    "master",
				},
			},
			desiredState: map[string]BucketObjectState{
				"master": {
					Body:   "master-body",
					Bucket: "mybucket",
					Key:    "master",
				},
			},
			expectedState: map[string]BucketObjectState{
				"master": {
					Body:   "master-body",
					Bucket: "mybucket",
					Key:    "master",
				},
			},
		},
		{
			description: "current state has 1 object, desired state has 2 objects, create change == missing object",
			obj: &v1alpha1.Cluster{
				Status: v1alpha1.ClusterStatus{
					ProviderStatus: &runtime.RawExtension{
						Raw: []byte(`
							{
								"cluster": {
									"id": "5xchu"
								}
							}
						`),
					},
				},
			},
			currentState: map[string]BucketObjectState{
				"master": {
					Body:   "master-body",
					Bucket: "mybucket",
					Key:    "master",
				},
			},
			desiredState: map[string]BucketObjectState{
				"master": {
					Body:   "master-body",
					Bucket: "mybucket",
					Key:    "master",
				},
				"worker": {
					Body:   "worker-body",
					Bucket: "mybucket",
					Key:    "worker",
				},
			},
			expectedState: map[string]BucketObjectState{
				"master": {},
				"worker": {
					Body:   "worker-body",
					Bucket: "mybucket",
					Key:    "worker",
				},
			},
		},
		{
			description: "current state matches desired state, empty create change",
			obj: &v1alpha1.Cluster{
				Status: v1alpha1.ClusterStatus{
					ProviderStatus: &runtime.RawExtension{
						Raw: []byte(`
							{
								"cluster": {
									"id": "5xchu"
								}
							}
						`),
					},
				},
			},
			currentState: map[string]BucketObjectState{
				"master": {
					Body:   "master-body",
					Bucket: "mybucket",
					Key:    "master",
				},
				"worker": {
					Body:   "worker-body",
					Bucket: "mybucket",
					Key:    "worker",
				},
			},
			desiredState: map[string]BucketObjectState{
				"master": {
					Body:   "master-body",
					Bucket: "mybucket",
					Key:    "master",
				},
				"worker": {
					Body:   "worker-body",
					Bucket: "mybucket",
					Key:    "worker",
				},
			},
			expectedState: map[string]BucketObjectState{
				"master": {},
				"worker": {},
			},
		},
	}

	awsClients := aws.Clients{
		S3: &S3ClientMock{},
	}
	cloudConfig := &CloudConfigMock{}

	var err error
	var newResource *Resource
	{
		c := Config{
			CertsSearcher:      certstest.NewSearcher(certstest.Config{}),
			CloudConfig:        cloudConfig,
			Logger:             microloggertest.New(),
			RandomKeysSearcher: randomkeystest.NewSearcher(),
		}

		newResource, err = New(c)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.TODO()
			cc := controllercontext.Context{
				Client: controllercontext.ContextClient{
					TenantCluster: controllercontext.ContextClientTenantCluster{
						AWS: awsClients,
					},
				},
			}
			ctx = controllercontext.NewContext(ctx, cc)

			result, err := newResource.newCreateChange(ctx, tc.obj, tc.currentState, tc.desiredState)
			if err != nil {
				t.Errorf("expected '%v' got '%#v'", nil, err)
			}
			createChange, ok := result.(map[string]BucketObjectState)
			if !ok {
				t.Errorf("expected '%T', got '%T'", createChange, result)
			}

			if !reflect.DeepEqual(tc.expectedState, createChange) {
				t.Error("expected", tc.expectedState, "got", createChange)
			}
		})
	}
}