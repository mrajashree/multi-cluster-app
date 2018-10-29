package v3

import (
	"github.com/rancher/norman/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GlobalDNS struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object’s metadata. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata,omitempty"`

	DNSName      string `json:"dnsName,omitempty" norman:"required"`
	RootDomain   string `json:"rootDomain" norman:"required"`
	TTLSeconds   int64  `json:"ttl" norman:"default=300"`
	ProviderName string `json:"providerName,omitempty" norman:"required"`
}

type MultiClusterApp struct {
	types.Namespaced

	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MultiClusterAppSpec   `json:"spec,omitempty"`
	Status MultiClusterAppStatus `json:"status,omitempty"`
}

type MultiClusterAppSpec struct {
	ChartRepositoryURL string   `json:"chartRepositoryUrl,omitempty"`
	ChartReference     string   `json:"chartReference,omitempty"`
	ChartVersion       string   `json:"chartVersion,omitempty"`
	ReleaseName        string   `json:"releaseName,omitempty"`
	ReleaseNamespace   string   `json:"releaseNamespace,omitempty"`
	Targets            []Target `json:"targets,omitempty" norman:"required"`
}

type MultiClusterAppStatus struct {
	HealthState string `json:"healthState,omitempty" norman:"required"`
}

type Target struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TargetSpec   `json:"spec,omitempty" norman:"required"`
	Status TargetStatus `json:"status,omitempty" norman:"required"`
}

type TargetSpec struct {
	ClusterConfig ClusterConfig     `json:"clusterConfig,omitempty" norman:"required"`
	Answers       map[string]string `json:"answers,omitempty"`
}

type TargetStatus struct {
	ChartReleaseName string `json:"chartReleaseName,omitempty" norman:"required"`
	HealthState      string `json:"healthState,omitempty" norman:"required"`
}

type ClusterConfig struct {
	Namespace                string `json:"namespace,omitempty"`
	Server                   string `json:"server,omitempty" norman:"required"`
	CertificateAuthorityPath string `json:"certificateAuthorityPath, omitempty"`
	ClientCertificatePath    string `json:"clientCertificatePath,omitempty"`
	ClientKeyPath            string `json:"clientKeyPath,omitempty"`
	TokenFile                string `json:"tokenFile,omitempty"`
}
