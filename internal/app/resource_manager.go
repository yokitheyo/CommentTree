package app

import (
	"github.com/wb-go/wbf/zlog"
)

type resourceManager struct {
	resources []resource
}

type resource struct {
	name      string
	closeFunc func() error
}

func (rm *resourceManager) addResource(res resource) {
	rm.resources = append(rm.resources, res)
}

func (rm *resourceManager) closeAll() error {
	var lastErr error
	for i := len(rm.resources) - 1; i >= 0; i-- {
		if err := rm.resources[i].closeFunc(); err != nil {
			zlog.Logger.Error().Err(err).Str("resource", rm.resources[i].name).Msg("failed to close resource")
			lastErr = err
		} else {
			zlog.Logger.Debug().Str("resource", rm.resources[i].name).Msg("resource closed")
		}
	}
	return lastErr
}
