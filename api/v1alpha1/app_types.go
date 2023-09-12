package v1alpha1

import (
	"go.uber.org/zap"
)

// Context TODO
type Context struct {
	Config *ConfigSpec
	Logger *zap.SugaredLogger
}
