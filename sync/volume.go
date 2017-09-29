package sync

import (
	"fmt"

	"github.com/rancher/go-rancher/v3"
	"github.com/rancher/netes-agent/utils"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
	storagev1 "k8s.io/client-go/pkg/apis/storage/v1"
)

const (
	nodeAffinityAnnotation      = "volume.alpha.kubernetes.io/node-affinity"
	nodeAffinityAnnotationValue = `
              {
                "nodeSelectorTerms": [
                    { "matchExpressions": [
                        { "key": "kubernetes.io/hostname",
                          "operator": "In",
                          "values": ["%s"]
                        }
                    ]}
                 ]}
              }`
)

func pvFromVolume(volume client.Volume, nodeName string) (v1.PersistentVolume, error) {
	name := utils.Hash(volume.Id)
	source, err := getVolumeSource(volume)
	if err != nil {
		return v1.PersistentVolume{}, err
	}

	pv := v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.PersistentVolumeSpec{
			StorageClassName: name,
			AccessModes: []v1.PersistentVolumeAccessMode{
				getAccessModeFromSource(source),
			},
			PersistentVolumeSource: source,
		},
	}

	if volume.SizeMb != 0 {
		size, err := getSize(volume)
		if err != nil {
			return v1.PersistentVolume{}, err
		}
		pv.Spec.Capacity = v1.ResourceList{
			"storage": size,
		}
	}

	if source.Local != nil && nodeName != "" {
		pv.Annotations = map[string]string{
			nodeAffinityAnnotation: fmt.Sprintf(nodeAffinityAnnotationValue, nodeName),
		}
	}

	return pv, nil
}

func pvcFromVolume(volume client.Volume, storageClass *storagev1.StorageClass, namespace string) (v1.PersistentVolumeClaim, error) {
	name := utils.Hash(volume.Id)

	claim := v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			StorageClassName: &name,
		},
	}

	if storageClass == nil {
		source, err := getVolumeSource(volume)
		if err != nil {
			return v1.PersistentVolumeClaim{}, err
		}
		claim.Spec.AccessModes = []v1.PersistentVolumeAccessMode{
			getAccessModeFromSource(source),
		}
	} else {
		claim.Spec.AccessModes = []v1.PersistentVolumeAccessMode{
			getAccessModeFromProvisioner(storageClass),
		}
	}

	if volume.SizeMb != 0 {
		size, err := getSize(volume)
		if err != nil {
			return v1.PersistentVolumeClaim{}, err
		}
		claim.Spec.Resources = v1.ResourceRequirements{
			Requests: v1.ResourceList{
				"storage": size,
			},
		}
	}

	if volume.StorageClass != "" {
		claim.Spec.StorageClassName = &volume.StorageClass
	}

	return claim, nil
}

func getSize(volume client.Volume) (resource.Quantity, error) {
	return resource.ParseQuantity(fmt.Sprintf("%dMi", volume.SizeMb))
}

func getVolumeSource(volume client.Volume) (v1.PersistentVolumeSource, error) {
	var source v1.PersistentVolumeSource
	err := utils.ConvertByJSON(volume.PvConfig, &source)
	return source, err
}

func getAccessModeFromSource(source v1.PersistentVolumeSource) v1.PersistentVolumeAccessMode {
	switch {
	case source.AzureFile != nil:
		fallthrough
	case source.CephFS != nil:
		fallthrough
	case source.Glusterfs != nil:
		fallthrough
	case source.Quobyte != nil:
		fallthrough
	case source.NFS != nil:
		fallthrough
	case source.PortworxVolume != nil:
		return v1.ReadWriteMany
	}
	return v1.ReadWriteOnce
}

func getAccessModeFromProvisioner(storageClass *storagev1.StorageClass) v1.PersistentVolumeAccessMode {
	switch storageClass.Provisioner {
	case "kubernetes.io/azure-file":
		fallthrough
	case "kubernetes.io/glusterfs":
		fallthrough
	case "kubernetes.io/quobyte":
		fallthrough
	case "kubernetes.io/portworx-volume":
		return v1.ReadWriteMany
	}
	return v1.ReadWriteOnce
}
