package discovery

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestNodeMatchesPV(t *testing.T) {
	pv := &corev1.PersistentVolume{
		Spec: corev1.PersistentVolumeSpec{
			NodeAffinity: &corev1.VolumeNodeAffinity{
				Required: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      "kubernetes.io/hostname",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{"node-a", "node-b"},
								},
							},
						},
					},
				},
			},
		},
	}

	if !nodeMatchesPV(pv, "node-a") {
		t.Error("expected node-a to match")
	}
	if nodeMatchesPV(pv, "node-c") {
		t.Error("expected node-c not to match")
	}
}

func TestResolveStorageClass(t *testing.T) {
	sc := "fast-ssd"
	pv := &corev1.PersistentVolume{
		Spec: corev1.PersistentVolumeSpec{
			StorageClassName: "pv-sc",
		},
	}
	pvc := &corev1.PersistentVolumeClaim{
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &sc,
		},
	}

	got := resolveStorageClass(pv, pvc, nil)
	if got != "pv-sc" {
		t.Errorf("resolveStorageClass = %q, want pv-sc", got)
	}
}

func TestTermMatchesNodeMetadataName(t *testing.T) {
	term := corev1.NodeSelectorTerm{
		MatchFields: []corev1.NodeSelectorRequirement{
			{
				Key:      "metadata.name",
				Operator: corev1.NodeSelectorOpIn,
				Values:   []string{"worker-1"},
			},
		},
	}

	if !termMatchesNode(term, "worker-1") {
		t.Error("expected worker-1 to match metadata.name selector")
	}
}
