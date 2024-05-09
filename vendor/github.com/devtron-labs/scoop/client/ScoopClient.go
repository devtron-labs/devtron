package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s"
	types2 "github.com/devtron-labs/scoop/types"
	"github.com/devtron-labs/scoop/utils"
	"go.uber.org/zap"
	"net/http"
)

type ScoopClient interface {
	GetApiResources(ctx context.Context) ([]*k8s.K8sApiResource, error)
	GetResourceList(ctx context.Context, resourceIdentifier *types2.K8sRequestBean) (*k8s.ClusterResourceListMap, error)
	UpdateK8sCacheConfig(ctx context.Context, config *types2.Config) error
	UpdateWatcherConfig(ctx context.Context, action types2.Action, watcher *types2.Watcher) error
	UpdateNamespaceConfig(ctx context.Context, action types2.Action, namespace string, isProd bool) error
}

type ScoopClientImpl struct {
	logger            *zap.SugaredLogger
	serverUrlWithPort string
	passKey           string
	client            http.Client
}

func NewScoopClientImpl(logger *zap.SugaredLogger, serverUrl string, passKey string) (ScoopClient, error) {
	client := http.Client{}
	return &ScoopClientImpl{
		logger:            logger,
		serverUrlWithPort: serverUrl,
		passKey:           passKey,
		client:            client,
	}, nil
}

func (impl *ScoopClientImpl) GetApiResources(ctx context.Context) ([]*k8s.K8sApiResource, error) {
	apiResourceUrl := impl.serverUrlWithPort + types2.API_RESOURCES_URL
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiResourceUrl, nil)
	if err != nil {
		impl.logger.Errorw("error while writing event", "err", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-PASS-KEY", impl.passKey)

	resp, err := impl.client.Do(req)
	if err != nil {
		impl.logger.Errorw("error in sending scoop rest request ", "err", err)
		return nil, err
	}
	impl.logger.Infof("scoop api response for api resources %s", resp.Status)
	defer resp.Body.Close()
	body := resp.Body
	decoder := json.NewDecoder(body)
	response := make([]*k8s.K8sApiResource, 0)
	err = decoder.Decode(&response)
	if err != nil {
		impl.logger.Errorw("error occurred while decoding scoop response", "err", err)
		return nil, err
	}
	return response, nil
}

func (impl *ScoopClientImpl) GetResourceList(ctx context.Context, k8sRequestBean *types2.K8sRequestBean) (*k8s.ClusterResourceListMap, error) {
	requestJson, err := json.Marshal(k8sRequestBean)
	resourceListUrl := impl.serverUrlWithPort + types2.RESOURCE_LIST_URL
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, resourceListUrl, bytes.NewBuffer(requestJson))
	if err != nil {
		impl.logger.Errorw("error while writing event", "err", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-PASS-KEY", impl.passKey)

	resp, err := impl.client.Do(req)
	if err != nil {
		impl.logger.Errorw("error in sending scoop rest request ", "err", err)
		return nil, err
	}
	impl.logger.Infof("scoop api response for get resource list %s", resp.Status)
	defer resp.Body.Close()
	body := resp.Body
	decoder := json.NewDecoder(body)

	var response k8s.ClusterResourceListMap
	err = decoder.Decode(&response)
	if err != nil {
		impl.logger.Errorw("error occurred while decoding scoop response", "err", err)
		return nil, err
	}
	return &response, nil
}

func (impl *ScoopClientImpl) UpdateK8sCacheConfig(ctx context.Context, config *types2.Config) error {
	requestJson, err := json.Marshal(config)
	cacheConfigUrl := impl.serverUrlWithPort + types2.K8S_CACHE_CONFIG_URL
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cacheConfigUrl, bytes.NewBuffer(requestJson))
	if err != nil {
		impl.logger.Errorw("error while writing event", "err", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-PASS-KEY", impl.passKey)

	resp, err := impl.client.Do(req)
	if err != nil {
		impl.logger.Errorw("error in sending scoop rest request ", "err", err)
		return err
	}
	impl.logger.Infof("scoop api response for get resource list %s", resp.Status)
	defer resp.Body.Close()
	return nil
}

func (impl *ScoopClientImpl) UpdateWatcherConfig(ctx context.Context, action types2.Action, watcher *types2.Watcher) error {
	if watcher == nil {
		return fmt.Errorf("watcher cannot be nil")
	}
	payload := types2.Payload{
		Action:  action,
		Watcher: watcher,
	}

	headers := map[string]string{
		"X-PASS-KEY": impl.passKey,
	}

	resp := &utils.Response{}

	err := utils.CallPostApi(impl.serverUrlWithPort+types2.WATCHER_CUD_URL, nil, headers, payload, resp)
	return err
}

func (impl *ScoopClientImpl) UpdateNamespaceConfig(ctx context.Context, action types2.Action, namespace string, isProd bool) error {
	if namespace == "" {
		return fmt.Errorf("namespace cannot be emepty")
	}
	headers := map[string]string{
		"X-PASS-KEY": impl.passKey,
	}

	isProdStr := "false"
	if isProd {
		isProdStr = "true"
	}
	queryParams := map[string]string{
		types2.NamespaceKey: namespace,
		types2.ActionKey:    string(action),
		types2.IsProdKey:    isProdStr,
	}

	resp := &utils.Response{}
	err := utils.CallPostApi(impl.serverUrlWithPort+types2.NAMESPACE_CUD_URL, queryParams, headers, map[string]string{}, resp)
	return err
}
