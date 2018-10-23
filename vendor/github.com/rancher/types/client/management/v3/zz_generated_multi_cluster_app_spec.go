package client

const (
	MultiClusterAppSpecType                    = "multiClusterAppSpec"
	MultiClusterAppSpecFieldChartReference     = "chartReference"
	MultiClusterAppSpecFieldChartRepositoryURL = "chartRepoUrl"
	MultiClusterAppSpecFieldChartVersion       = "chartVersion"
	MultiClusterAppSpecFieldReleaseName        = "releaseName"
	MultiClusterAppSpecFieldReleaseNamespace   = "releaseNamespace"
	MultiClusterAppSpecFieldTargets            = "targets"
)

type MultiClusterAppSpec struct {
	ChartReference     string   `json:"chartReference,omitempty" yaml:"chartReference,omitempty"`
	ChartRepositoryURL string   `json:"chartRepoUrl,omitempty" yaml:"chartRepoUrl,omitempty"`
	ChartVersion       string   `json:"chartVersion,omitempty" yaml:"chartVersion,omitempty"`
	ReleaseName        string   `json:"releaseName,omitempty" yaml:"releaseName,omitempty"`
	ReleaseNamespace   string   `json:"releaseNamespace,omitempty" yaml:"releaseNamespace,omitempty"`
	Targets            []Target `json:"targets,omitempty" yaml:"targets,omitempty"`
}
