package multiclusterapp

import (
	"github.com/rancher/norman/api/handler"
	"github.com/rancher/norman/types"
	"github.com/sirupsen/logrus"
)

func MCAppListHandler(request *types.APIContext, next types.RequestHandler) error {
	logrus.Debugf("MCAppListHandler called")

	return handler.ListHandler(request, next)
}