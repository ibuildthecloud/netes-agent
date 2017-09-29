package sync

import (
	log "github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher/v3"
	"github.com/rancher/netes-agent/watch"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	storagev1 "k8s.io/client-go/pkg/apis/storage/v1"
)

func reconcileVolumes(clientset *kubernetes.Clientset, watchClient *watch.Client, deploymentUnit client.DeploymentSyncRequest) error {
	for _, volume := range deploymentUnit.Volumes {
		var storageClass *storagev1.StorageClass
		var err error
		if volume.StorageClass == "" {
			pv, err := pvFromVolume(volume, deploymentUnit.NodeName)
			if err != nil {
				return err
			}
			if err := createPv(clientset, pv); err != nil {
				return err
			}
		} else {
			storageClass, err = clientset.StorageV1().StorageClasses().Get(volume.StorageClass, metav1.GetOptions{})
			if err != nil {
				return err
			}
		}

		pvc, err := pvcFromVolume(volume, storageClass, deploymentUnit.Namespace)
		if err != nil {
			return err
		}
		if err := createPvc(clientset, pvc); err != nil {
			return err
		}
	}
	return nil
}

func createPv(clientset *kubernetes.Clientset, pv v1.PersistentVolume) error {
	log.Infof("Creating persistent volume %s", pv.Name)
	_, err := clientset.PersistentVolumes().Create(&pv)
	if errors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func createPvc(clientset *kubernetes.Clientset, pvc v1.PersistentVolumeClaim) error {
	log.Infof("Creating persistent volume claim %s", pvc.Name)
	_, err := clientset.PersistentVolumeClaims(v1.NamespaceDefault).Create(&pvc)
	if errors.IsAlreadyExists(err) {
		return nil
	}
	return err
}
