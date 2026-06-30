package discovery

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/alvarorg/local-pvc-exporter/internal/kube"
)

// VolumeType identifies the PV volume source type.
type VolumeType string

const (
	VolumeTypeHostPath VolumeType = "hostpath"
	VolumeTypeLocal    VolumeType = "local"
)

// Volume represents a discovered PVC-backed volume on this node.
type Volume struct {
	PVCName       string
	Namespace     string
	PVName        string
	StorageClass  string
	NodeName      string
	VolumeType    VolumeType
	HostPath      string
	CapacityBytes int64
}

// Discoverer finds hostPath and local PVs bound to PVCs on the current node.
type Discoverer struct {
	client   *kube.Client
	nodeName string
	hostRoot string
}

// New creates a new Discoverer.
func New(client *kube.Client, nodeName, hostRoot string) *Discoverer {
	return &Discoverer{
		client:   client,
		nodeName: nodeName,
		hostRoot: hostRoot,
	}
}

// Discover returns volumes that should be measured on this node.
func (d *Discoverer) Discover(ctx context.Context) ([]Volume, error) {
	pvList, err := d.client.ListPersistentVolumes(ctx)
	if err != nil {
		return nil, fmt.Errorf("list persistent volumes: %w", err)
	}

	pvcList, err := d.client.ListPersistentVolumeClaims(ctx)
	if err != nil {
		return nil, fmt.Errorf("list persistent volume claims: %w", err)
	}

	scList, err := d.client.ListStorageClasses(ctx)
	if err != nil {
		return nil, fmt.Errorf("list storage classes: %w", err)
	}

	pvcByPV := make(map[string]*corev1.PersistentVolumeClaim, len(pvcList.Items))
	for i := range pvcList.Items {
		pvc := &pvcList.Items[i]
		if pvc.Spec.VolumeName != "" {
			pvcByPV[pvc.Spec.VolumeName] = pvc
		}
	}

	scNames := make(map[string]struct{}, len(scList.Items))
	for _, sc := range scList.Items {
		scNames[sc.Name] = struct{}{}
	}

	var volumes []Volume
	for i := range pvList.Items {
		pv := &pvList.Items[i]
		if pv.Status.Phase != corev1.VolumeBound {
			continue
		}

		pvc, ok := pvcByPV[pv.Name]
		if !ok {
			continue
		}

		vol, ok, err := d.volumeFromPV(pv, pvc, scNames)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		volumes = append(volumes, vol)
	}

	return volumes, nil
}

func (d *Discoverer) volumeFromPV(
	pv *corev1.PersistentVolume,
	pvc *corev1.PersistentVolumeClaim,
	scNames map[string]struct{},
) (Volume, bool, error) {
	capacity := pv.Spec.Capacity[corev1.ResourceStorage]
	capacityBytes := capacity.Value()

	storageClass := resolveStorageClass(pv, pvc, scNames)

	switch {
	case pv.Spec.HostPath != nil:
		hostPath := pv.Spec.HostPath.Path
		if hostPath == "" {
			return Volume{}, false, nil
		}
		resolved := filepath.Join(d.hostRoot, strings.TrimPrefix(hostPath, "/"))
		if !pathExists(resolved) {
			return Volume{}, false, nil
		}
		return Volume{
			PVCName:       pvc.Name,
			Namespace:     pvc.Namespace,
			PVName:        pv.Name,
			StorageClass:  storageClass,
			NodeName:      d.nodeName,
			VolumeType:    VolumeTypeHostPath,
			HostPath:      resolved,
			CapacityBytes: capacityBytes,
		}, true, nil

	case pv.Spec.Local != nil:
		if !nodeMatchesPV(pv, d.nodeName) {
			return Volume{}, false, nil
		}
		localPath := pv.Spec.Local.Path
		if localPath == "" {
			return Volume{}, false, nil
		}
		resolved := filepath.Join(d.hostRoot, strings.TrimPrefix(localPath, "/"))
		if !pathExists(resolved) {
			return Volume{}, false, nil
		}
		return Volume{
			PVCName:       pvc.Name,
			Namespace:     pvc.Namespace,
			PVName:        pv.Name,
			StorageClass:  storageClass,
			NodeName:      d.nodeName,
			VolumeType:    VolumeTypeLocal,
			HostPath:      resolved,
			CapacityBytes: capacityBytes,
		}, true, nil

	default:
		return Volume{}, false, nil
	}
}

func resolveStorageClass(pv *corev1.PersistentVolume, pvc *corev1.PersistentVolumeClaim, scNames map[string]struct{}) string {
	if pv.Spec.StorageClassName != "" {
		return pv.Spec.StorageClassName
	}
	if pvc.Spec.StorageClassName != nil && *pvc.Spec.StorageClassName != "" {
		return *pvc.Spec.StorageClassName
	}
	if ann, ok := pv.Annotations["volume.beta.kubernetes.io/storage-class"]; ok && ann != "" {
		return ann
	}
	if ann, ok := pvc.Annotations["volume.beta.kubernetes.io/storage-class"]; ok && ann != "" {
		return ann
	}
	_ = scNames
	return ""
}

func nodeMatchesPV(pv *corev1.PersistentVolume, nodeName string) bool {
	if pv.Spec.NodeAffinity == nil || pv.Spec.NodeAffinity.Required == nil {
		return false
	}

	for _, term := range pv.Spec.NodeAffinity.Required.NodeSelectorTerms {
		if termMatchesNode(term, nodeName) {
			return true
		}
	}
	return false
}

func termMatchesNode(term corev1.NodeSelectorTerm, nodeName string) bool {
	for _, expr := range term.MatchExpressions {
		if expr.Key != "kubernetes.io/hostname" {
			continue
		}
		switch expr.Operator {
		case corev1.NodeSelectorOpIn:
			for _, v := range expr.Values {
				if v == nodeName {
					return true
				}
			}
		case corev1.NodeSelectorOpExists:
			return true
		}
	}

	for _, field := range term.MatchFields {
		if field.Key == "metadata.name" && field.Operator == corev1.NodeSelectorOpIn {
			for _, v := range field.Values {
				if v == nodeName {
					return true
				}
			}
		}
	}

	return false
}

func pathExists(path string) bool {
	return fileExists(path)
}
