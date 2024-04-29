package client

import (
	"fmt"
	"github.com/devtron-labs/scoop/pkg/watcherEvents"
	"github.com/devtron-labs/scoop/utils"
)

func UpdateWatcherConfig(serviceUrl, passKey string, action watcherEvents.Action, watcher *watcherEvents.Watcher) error {
	if watcher == nil {
		return fmt.Errorf("watcher cannot be nil")
	}
	payload := watcherEvents.Payload{
		Action:  action,
		Watcher: watcher,
	}

	headers := map[string]string{
		"X-PASS-KEY": passKey,
	}

	resp := &utils.Response{}

	err := utils.CallPostApi(serviceUrl+watcherEvents.WATCHER_CUD_URL, nil, headers, payload, resp)
	return err
}
