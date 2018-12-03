package adapter

import (
	"encoding/base64"
	"reflect"
	"strings"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func TestAdapterLaunchConfigurationRegularFields(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		description                      string
		customObject                     v1alpha1.AWSConfig
		expectedError                    bool
		expectedInstanceType             string
		expectedAssociatePublicIPAddress bool
		expectedBlockDeviceMappings      []BlockDeviceMapping
	}{
		{
			description: "basic matching, all fields present",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
					},
					AWS: v1alpha1.AWSConfigSpecAWS{
						Workers: []v1alpha1.AWSConfigSpecAWSNode{
							{
								InstanceType: "myinstancetype",
							},
						},
					},
				},
			},
			expectedInstanceType:             "myinstancetype",
			expectedAssociatePublicIPAddress: false,
			expectedBlockDeviceMappings: []BlockDeviceMapping{
				{
					DeleteOnTermination: true,
					DeviceName:          defaultEBSVolumeMountPoint,
					VolumeSize:          defaultEBSVolumeSize,
					VolumeType:          defaultEBSVolumeType,
				},
			},
		},
	}
	for _, tc := range testCases {
		clients := Clients{
			EC2: &EC2ClientMock{},
			IAM: &IAMClientMock{},
			STS: &STSClientMock{},
		}
		a := Adapter{}

		t.Run(tc.description, func(t *testing.T) {
			cfg := Config{
				CustomObject: tc.customObject,
				Clients:      clients,
			}
			err := a.Guest.LaunchConfiguration.Adapt(cfg)
			if tc.expectedError && err == nil {
				t.Error("expected error didn't happen")
			}

			if !tc.expectedError && err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.Guest.LaunchConfiguration.ASGType != prefixWorker {
				t.Errorf("unexpected ASGType, got %q, want %q", a.Guest.LaunchConfiguration.ASGType, prefixWorker)
			}
			if a.Guest.LaunchConfiguration.WorkerInstanceType != tc.expectedInstanceType {
				t.Errorf("unexpected InstanceType, got %q, want %q", a.Guest.LaunchConfiguration.WorkerInstanceType, tc.expectedInstanceType)
			}
			if a.Guest.LaunchConfiguration.WorkerAssociatePublicIPAddress != tc.expectedAssociatePublicIPAddress {
				t.Errorf("unexpected WorkerAssociatePublicIPAddress, got %t, want %t", a.Guest.LaunchConfiguration.WorkerAssociatePublicIPAddress, tc.expectedAssociatePublicIPAddress)
			}
			if !reflect.DeepEqual(a.Guest.LaunchConfiguration.WorkerBlockDeviceMappings, tc.expectedBlockDeviceMappings) {
				t.Errorf("unexpected BlockDeviceMappings, got %v, want %v", a.Guest.LaunchConfiguration.WorkerBlockDeviceMappings, tc.expectedBlockDeviceMappings)
			}
		})
	}
}

func TestAdapterLaunchConfigurationSmallCloudConfig(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		description  string
		expectedLine string
	}{
		{
			description:  "userdata file",
			expectedLine: "USERDATA_FILE=worker",
		},
	}

	a := Adapter{}
	clients := Clients{
		EC2: &EC2ClientMock{},
		IAM: &IAMClientMock{},
		STS: &STSClientMock{accountID: "000000000000"},
	}
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
			AWS: v1alpha1.AWSConfigSpecAWS{
				Region: "myregion",
				Workers: []v1alpha1.AWSConfigSpecAWSNode{
					{
						ImageID:      "myimageid",
						InstanceType: "myinstancetype",
					},
				},
			},
		},
	}
	cfg := Config{
		CustomObject: customObject,
		Clients:      clients,
	}
	err := a.Guest.LaunchConfiguration.Adapt(cfg)

	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	data, err := base64.StdEncoding.DecodeString(a.Guest.LaunchConfiguration.WorkerSmallCloudConfig)
	if err != nil {
		t.Errorf("unexpected error decoding SmallCloudConfig %v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			if !strings.Contains(string(data), tc.expectedLine) {
				t.Errorf("SmallCloudConfig didn't contain expected %q, complete: %q", tc.expectedLine, string(data))
			}
		})
	}
}