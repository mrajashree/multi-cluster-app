package controllers

import (
	"context"

	"github.com/rancher/multi-cluster-app/controllers/globaldns"
	"github.com/rancher/multi-cluster-app/controllers/multiclusterapp"
	"github.com/rancher/types/config"
)

func Register(ctx context.Context, management *config.ManagementContext) {
	globaldns.Register(ctx, management)
	multiclusterapp.Register(ctx, management)
}
