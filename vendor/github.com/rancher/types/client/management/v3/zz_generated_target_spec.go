package client

const (
	TargetSpecType               = "targetSpec"
	TargetSpecFieldAnswers       = "answers"
	TargetSpecFieldClusterConfig = "clusterConfig"
)

type TargetSpec struct {
	Answers       map[string]string `json:"answers,omitempty" yaml:"answers,omitempty"`
	ClusterConfig *ClusterConfig    `json:"clusterConfig,omitempty" yaml:"clusterConfig,omitempty"`
}
