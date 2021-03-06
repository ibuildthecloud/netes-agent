package sync

import (
	"testing"

	"github.com/rancher/go-rancher/v3"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/pkg/api/v1"
)

func TestGetPodSpec(t *testing.T) {
	assert.Equal(t, getPodSpec(client.DeploymentSyncRequest{
		Containers: []client.Container{
			client.Container{},
		},
	}), v1.PodSpec{
		RestartPolicy: v1.RestartPolicyNever,
		DNSPolicy:     v1.DNSDefault,
	})

	assert.Equal(t, getPodSpec(client.DeploymentSyncRequest{
		Containers: []client.Container{
			client.Container{
				RestartPolicy: &client.RestartPolicy{
					Name: "always",
				},
				PrimaryNetworkId: "1",
				IpcMode:          "host",
				PidMode:          "host",
			},
		},
		Networks: []client.Network{
			client.Network{
				Resource: client.Resource{
					Id: "1",
				},
				Kind: hostNetworkingKind,
			},
		},
	}), v1.PodSpec{
		RestartPolicy: v1.RestartPolicyNever,
		HostIPC:       true,
		HostNetwork:   true,
		HostPID:       true,
		DNSPolicy:     v1.DNSDefault,
	})
}

func TestSecurityContext(t *testing.T) {
	securityContext := getSecurityContext(client.Container{
		Privileged: true,
		ReadOnly:   true,
		CapAdd: []string{
			"capadd1",
			"capadd2",
		},
		CapDrop: []string{
			"capdrop1",
			"capdrop2",
		},
	})
	assert.Equal(t, *securityContext.Privileged, true)
	assert.Equal(t, *securityContext.ReadOnlyRootFilesystem, true)
	assert.Equal(t, securityContext.Capabilities.Add, []v1.Capability{
		v1.Capability("capadd1"),
		v1.Capability("capadd2"),
	})
	assert.Equal(t, securityContext.Capabilities.Drop, []v1.Capability{
		v1.Capability("capdrop1"),
		v1.Capability("capdrop2"),
	})
}

func TestGetVolumes(t *testing.T) {
	assert.Equal(t, getVolumes(client.DeploymentSyncRequest{
		Containers: []client.Container{
			client.Container{
				DataVolumes: []string{
					"/host/path:/container/path",
				},
			},
		},
	}), []v1.Volume{
		v1.Volume{
			Name: "host-path-volume",
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: "/host/path",
				},
			},
		},
	})
	assert.Equal(t, len(getVolumes(client.DeploymentSyncRequest{
		Containers: []client.Container{
			client.Container{
				DataVolumes: []string{
					"/anonymous/volume",
				},
			},
		},
	})), 0)
}

func TestGetVolumeMounts(t *testing.T) {
	assert.Equal(t, getVolumeMounts(client.Container{
		DataVolumes: []string{
			"/host/path:/container/path",
		},
	}), []v1.VolumeMount{
		v1.VolumeMount{
			Name:      "host-path-volume",
			MountPath: "/container/path",
		},
	})
	assert.Equal(t, len(getVolumeMounts(client.Container{
		DataVolumes: []string{
			"/anonymous/volume",
		},
	})), 0)
}
