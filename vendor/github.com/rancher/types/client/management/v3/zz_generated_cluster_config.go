package client

const (
	ClusterConfigType                          = "clusterConfig"
	ClusterConfigFieldCertificateAuthorityPath = "certificateAuthorityPath"
	ClusterConfigFieldClientCertificatePath    = "clientCertificatePath"
	ClusterConfigFieldClientKeyPath            = "clientKeyPath"
	ClusterConfigFieldNamespace                = "namespace"
	ClusterConfigFieldServer                   = "server"
	ClusterConfigFieldTokenFile                = "tokenFile"
)

type ClusterConfig struct {
	CertificateAuthorityPath string `json:"certificateAuthorityPath,omitempty" yaml:"certificateAuthorityPath,omitempty"`
	ClientCertificatePath    string `json:"clientCertificatePath,omitempty" yaml:"clientCertificatePath,omitempty"`
	ClientKeyPath            string `json:"clientKeyPath,omitempty" yaml:"clientKeyPath,omitempty"`
	Namespace                string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Server                   string `json:"server,omitempty" yaml:"server,omitempty"`
	TokenFile                string `json:"tokenFile,omitempty" yaml:"tokenFile,omitempty"`
}
