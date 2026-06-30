package kube

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps the Kubernetes API client.
type Client struct {
	clientset kubernetes.Interface
}

// New creates a Kubernetes client from in-cluster config or kubeconfig.
func New(kubeconfig string) (*Client, error) {
	cfg, err := buildConfig(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("build kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("create clientset: %w", err)
	}

	return &Client{clientset: clientset}, nil
}

func buildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	if _, err := rest.InClusterConfig(); err == nil {
		return rest.InClusterConfig()
	}

	// Fall back to default kubeconfig location for local development.
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("in-cluster config unavailable and cannot resolve home dir: %w", err)
	}
	defaultPath := filepath.Join(home, ".kube", "config")
	return clientcmd.BuildConfigFromFlags("", defaultPath)
}

// ListPersistentVolumes returns all PVs.
func (c *Client) ListPersistentVolumes(ctx context.Context) (*corev1.PersistentVolumeList, error) {
	return c.clientset.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
}

// ListPersistentVolumeClaims returns all PVCs.
func (c *Client) ListPersistentVolumeClaims(ctx context.Context) (*corev1.PersistentVolumeClaimList, error) {
	return c.clientset.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
}

// ListStorageClasses returns all StorageClasses.
func (c *Client) ListStorageClasses(ctx context.Context) (*storagev1.StorageClassList, error) {
	return c.clientset.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
}

// GetNode returns a node by name.
func (c *Client) GetNode(ctx context.Context, name string) (*corev1.Node, error) {
	return c.clientset.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
}
