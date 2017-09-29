package sync

import (
	"fmt"
	"testing"

	"k8s.io/client-go/pkg/api/v1"

	"github.com/rancher/go-rancher/v3"
	"github.com/rancher/netes-agent/utils"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	storagev1 "k8s.io/client-go/pkg/apis/storage/v1"
)

func TestPvFromVolume(t *testing.T) {
	pv, err := pvFromVolume(client.Volume{
		Resource: client.Resource{
			Id: "id1",
		},
		Name:   "volume1",
		SizeMb: 1000,
		PvConfig: map[string]interface{}{
			"gcePersistentDisk": map[string]interface{}{
				"fsType": "ext4",
				"pdName": "volume1",
			},
		},
	}, "")
	assert.NoError(t, err)
	assert.Equal(t, v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: utils.Hash("id1"),
		},
		Spec: v1.PersistentVolumeSpec{
			StorageClassName: utils.Hash("id1"),
			Capacity: v1.ResourceList{
				"storage": resource.MustParse("1000Mi"),
			},
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				GCEPersistentDisk: &v1.GCEPersistentDiskVolumeSource{
					FSType: "ext4",
					PDName: "volume1",
				},
			},
		},
	}, pv)

	pv, err = pvFromVolume(client.Volume{
		Resource: client.Resource{
			Id: "id1",
		},
		Name:   "volume1",
		SizeMb: 1000,
		PvConfig: map[string]interface{}{
			"nfs": map[string]interface{}{
				"server": "0.0.0.0",
				"path":   "/",
			},
		},
	}, "")
	assert.NoError(t, err)
	assert.Equal(t, v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: utils.Hash("id1"),
		},
		Spec: v1.PersistentVolumeSpec{
			StorageClassName: utils.Hash("id1"),
			Capacity: v1.ResourceList{
				"storage": resource.MustParse("1000Mi"),
			},
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteMany,
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				NFS: &v1.NFSVolumeSource{
					Server: "0.0.0.0",
					Path:   "/",
				},
			},
		},
	}, pv)
}

func TestPvFromLocalVolume(t *testing.T) {
	pv, err := pvFromVolume(client.Volume{
		Resource: client.Resource{
			Id: "id1",
		},
		Name:   "volume1",
		SizeMb: 1000,
		PvConfig: map[string]interface{}{
			"local": map[string]interface{}{
				"path": "/mnt/disks/ssd1",
			},
		},
	}, "node1")
	assert.NoError(t, err)
	assert.Equal(t, v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: utils.Hash("id1"),
			Annotations: map[string]string{
				nodeAffinityAnnotation: fmt.Sprintf(nodeAffinityAnnotationValue, "node1"),
			},
		},
		Spec: v1.PersistentVolumeSpec{
			StorageClassName: utils.Hash("id1"),
			Capacity: v1.ResourceList{
				"storage": resource.MustParse("1000Mi"),
			},
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				Local: &v1.LocalVolumeSource{
					Path: "/mnt/disks/ssd1",
				},
			},
		},
	}, pv)
}

func TestPvcFromVolume(t *testing.T) {
	pvc, err := pvcFromVolume(client.Volume{
		Resource: client.Resource{
			Id: "id1",
		},
		Name:   "volume1",
		SizeMb: 1000,
		PvConfig: map[string]interface{}{
			"nfs": map[string]interface{}{
				"server": "0.0.0.0",
				"path":   "/",
			},
		},
	}, nil, "default")
	assert.NoError(t, err)
	assert.Equal(t, v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.Hash("id1"),
			Namespace: "default",
		},
		Spec: v1.PersistentVolumeClaimSpec{
			StorageClassName: &[]string{utils.Hash("id1")}[0],
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					"storage": resource.MustParse("1000Mi"),
				},
			},
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteMany,
			},
		},
	}, pvc)

	pvc, err = pvcFromVolume(client.Volume{
		Resource: client.Resource{
			Id: "id1",
		},
		Name:         "volume1",
		SizeMb:       1000,
		StorageClass: "class1",
	}, &storagev1.StorageClass{
		Provisioner: "kubernetes.io/gce-pd",
	}, "default")
	assert.NoError(t, err)
	assert.Equal(t, v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.Hash("id1"),
			Namespace: "default",
		},
		Spec: v1.PersistentVolumeClaimSpec{
			StorageClassName: &[]string{"class1"}[0],
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					"storage": resource.MustParse("1000Mi"),
				},
			},
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
		},
	}, pvc)

	pvc, err = pvcFromVolume(client.Volume{
		Resource: client.Resource{
			Id: "id1",
		},
		Name:         "volume1",
		SizeMb:       1000,
		StorageClass: "class1",
	}, &storagev1.StorageClass{
		Provisioner: "kubernetes.io/azure-file",
	}, "default")
	assert.NoError(t, err)
	assert.Equal(t, v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.Hash("id1"),
			Namespace: "default",
		},
		Spec: v1.PersistentVolumeClaimSpec{
			StorageClassName: &[]string{"class1"}[0],
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					"storage": resource.MustParse("1000Mi"),
				},
			},
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteMany,
			},
		},
	}, pvc)
}
