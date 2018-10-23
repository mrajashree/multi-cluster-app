package multiclusterapp

import (
	"strings"
	"context"

	"github.com/rancher/norman/types"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/rancher/types/config"
	"github.com/sirupsen/logrus"
)

const (
	multiClusterAppController = "mgmt-multicluster-app-controller"
)

type MultiClusterAppController struct {
	mcappClient v3.MultiClusterAppInterface
	mcappLister v3.MultiClusterAppLister
}

func newMultiClusterApp(mgmt *config.ManagementContext) *MultiClusterAppController {
	m := MultiClusterAppController{
		mcappClient: mgmt.Management.MultiClusterApps(""),
		mcappLister: mgmt.Management.MultiClusterApps("").Controller().Lister(),
	}

	return &m
}

func Register(ctx context.Context, management *config.ManagementContext) {
	m := newMultiClusterApp(management)
	management.Management.MultiClusterApps("").AddHandler(multiClusterAppController, m.sync)
}


func (m MultiClusterAppController) sync(key string, obj *v3.MultiClusterApp) error {
	return nil
}