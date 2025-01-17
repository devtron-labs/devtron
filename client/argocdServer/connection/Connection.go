/*
 * Copyright (c) 2020-2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package connection

import (
	"context"
	"fmt"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/account"
	"github.com/argoproj/argo-cd/v2/util/settings"
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/client/argocdServer/bean"
	config2 "github.com/devtron-labs/devtron/client/argocdServer/config"
	"github.com/devtron-labs/devtron/client/argocdServer/session"
	"github.com/devtron-labs/devtron/client/argocdServer/version"
	bean2 "github.com/devtron-labs/devtron/pkg/cluster/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	k8s2 "github.com/devtron-labs/devtron/pkg/k8s"
	moduleRepo "github.com/devtron-labs/devtron/pkg/module/repo"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"math/rand"
	"strconv"
	"strings"
)

func init() {
	grpc_prometheus.EnableClientHandlingTimeHistogram()
}

const (
	DEVTRON_USER                     = "devtron"
	DEVTRONCD_NAMESPACE              = "devtroncd"
	ARGOCD_CM                        = "argocd-cm"
	ARGOCD_SECRET                    = "argocd-secret"
	ARGO_USER_APIKEY_CAPABILITY      = "apiKey"
	ARGO_USER_LOGIN_CAPABILITY       = "login"
	DEVTRON_ARGOCD_USERNAME_KEY      = "DEVTRON_ACD_USER_NAME"
	DEVTRON_ARGOCD_USER_PASSWORD_KEY = "DEVTRON_ACD_USER_PASSWORD"
	DEVTRON_ARGOCD_TOKEN_KEY         = "DEVTRON_ACD_TOKEN"
)

type ArgoCDConnectionManager interface {
	GetGrpcClientConnection(grpcConfig *bean.ArgoGRPCConfig) *grpc.ClientConn
	GetOrUpdateArgoCdUserDetail(grpcConfig *bean.ArgoGRPCConfig) string
}
type ArgoCDConnectionManagerImpl struct {
	logger                  *zap.SugaredLogger
	settingsManager         *settings.SettingsManager
	moduleRepository        moduleRepo.ModuleRepository
	argoCDSettings          *settings.ArgoCDSettings
	devtronSecretConfig     *util2.DevtronSecretConfig
	k8sUtil                 *k8s.K8sServiceImpl
	k8sCommonService        k8s2.K8sCommonService
	versionService          version.VersionService
	gitOpsConfigReadService config.GitOpsConfigReadService
	runTimeConfig           *k8s.RuntimeConfig
	argoCDConfigGetter      config2.ArgoCDConfigGetter
}

func NewArgoCDConnectionManagerImpl(Logger *zap.SugaredLogger,
	settingsManager *settings.SettingsManager,
	moduleRepository moduleRepo.ModuleRepository,
	environmentVariables *util2.EnvironmentVariables,
	k8sUtil *k8s.K8sServiceImpl,
	k8sCommonService k8s2.K8sCommonService,
	versionService version.VersionService,
	gitOpsConfigReadService config.GitOpsConfigReadService,
	runTimeConfig *k8s.RuntimeConfig,
	argoCDConfigGetter config2.ArgoCDConfigGetter) (*ArgoCDConnectionManagerImpl, error) {
	argoUserServiceImpl := &ArgoCDConnectionManagerImpl{
		logger:                  Logger,
		settingsManager:         settingsManager,
		moduleRepository:        moduleRepository,
		argoCDSettings:          nil,
		devtronSecretConfig:     environmentVariables.DevtronSecretConfig,
		k8sUtil:                 k8sUtil,
		k8sCommonService:        k8sCommonService,
		versionService:          versionService,
		gitOpsConfigReadService: gitOpsConfigReadService,
		runTimeConfig:           runTimeConfig,
		argoCDConfigGetter:      argoCDConfigGetter,
	}
	if !runTimeConfig.LocalDevMode {
		grpcConfig, err := argoCDConfigGetter.GetGRPCConfig()
		if err != nil {
			Logger.Errorw("error in GetAllGRPCConfigs", "error", err)
		}
		go argoUserServiceImpl.ValidateGitOpsAndGetOrUpdateArgoCdUserDetail(grpcConfig)
	}
	return argoUserServiceImpl, nil
}

const (
	ModuleNameArgoCd      string = "argo-cd"
	ModuleStatusInstalled string = "installed"
)

func (impl *ArgoCDConnectionManagerImpl) ValidateGitOpsAndGetOrUpdateArgoCdUserDetail(grpcConfig *bean.ArgoGRPCConfig) string {
	gitOpsConfigurationStatus, err := impl.gitOpsConfigReadService.IsGitOpsConfigured()
	if err != nil || !gitOpsConfigurationStatus.IsGitOpsConfigured {
		return ""
	}
	_ = impl.GetOrUpdateArgoCdUserDetail(grpcConfig)
	return ""
}

// GetConnection - this function will call only for acd connection
func (impl *ArgoCDConnectionManagerImpl) GetGrpcClientConnection(grpcConfig *bean.ArgoGRPCConfig) *grpc.ClientConn {
	//TODO: acdAuthConfig should be passed as arg in function
	token, err := impl.getLatestDevtronArgoCdUserToken(grpcConfig)
	if err != nil {
		impl.logger.Errorw("error in getting latest devtron argocd user token", "err", err)
	}
	return impl.getConnectionWithToken(grpcConfig.ConnectionConfig, token)
}

func (impl *ArgoCDConnectionManagerImpl) getConnectionWithToken(connectionConfig *bean.Config, token string) *grpc.ClientConn {
	//TODO: config should be passed to this function as argument
	settings := impl.getArgoCdSettings()
	var option []grpc.DialOption
	option = append(option, grpc.WithTransportCredentials(GetTLS(settings.Certificate)))
	if len(token) > 0 {
		option = append(option, grpc.WithPerRPCCredentials(TokenAuth{token: token}))
	}
	option = append(option, grpc.WithChainUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor, otelgrpc.UnaryClientInterceptor()), grpc.WithChainStreamInterceptor(grpc_prometheus.StreamClientInterceptor, otelgrpc.StreamClientInterceptor()))
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", connectionConfig.Host, connectionConfig.Port), option...)
	if err != nil {
		return nil
	}
	return conn
}

func (impl *ArgoCDConnectionManagerImpl) getLatestDevtronArgoCdUserToken(grpcConfig *bean.ArgoGRPCConfig) (string, error) {
	var (
		k8sClient *v1.CoreV1Client
		err       error
	)
	authConfig := grpcConfig.AuthConfig
	if authConfig.ClusterId == bean2.DefaultClusterId {
		k8sClient, err = impl.k8sUtil.GetCoreV1ClientInCluster()
		if err != nil {
			impl.logger.Errorw("error in getting k8s client for default cluster", "err", err)
			return "", err
		}
	} else {
		_, k8sClient, err = impl.k8sCommonService.GetCoreClientByClusterId(authConfig.ClusterId)
		if err != nil {
			impl.logger.Errorw("error in getting k8s client for default cluster", "err", err)
		}
	}
	devtronSecret, err := getSecret(impl.devtronSecretConfig.DevtronDexSecretNamespace, impl.devtronSecretConfig.DevtronSecretName, k8sClient)
	if err != nil {
		impl.logger.Errorw("error in getting devtron secret", "err", err)
		return "", err
	}
	secretData := devtronSecret.Data
	username := secretData[DEVTRON_ARGOCD_USERNAME_KEY]
	password := secretData[DEVTRON_ARGOCD_USER_PASSWORD_KEY]
	latestTokenNo := 1
	var token string
	for key, value := range secretData {
		if strings.HasPrefix(key, DEVTRON_ARGOCD_TOKEN_KEY) {
			keySplits := strings.Split(key, "_")
			keyLen := len(keySplits)
			tokenNo, err := strconv.Atoi(keySplits[keyLen-1])
			if err != nil {
				impl.logger.Errorw("error in converting token no string to integer", "err", err, "tokenNoString", keySplits[keyLen-1])
				return "", err
			}
			if tokenNo > latestTokenNo {
				latestTokenNo = tokenNo
				token = string(value)
			}
		}
	}

	if len(token) == 0 {
		newTokenNo := latestTokenNo + 1
		token, err = impl.createNewArgoCdTokenForDevtron(grpcConfig.ConnectionConfig, string(username), string(password), newTokenNo, k8sClient)
		if err != nil {
			impl.logger.Errorw("error in creating new argo cd token for devtron", "err", err)
			return "", err
		}
	}
	return token, nil
}

func (impl *ArgoCDConnectionManagerImpl) GetOrUpdateArgoCdUserDetail(grpcConfig *bean.ArgoGRPCConfig) string {

	token := ""
	var (
		k8sClient *v1.CoreV1Client
		err       error
	)

	authConfig := grpcConfig.AuthConfig
	if authConfig.ClusterId == bean2.DefaultClusterId {
		k8sClient, err = impl.k8sUtil.GetCoreV1ClientInCluster()
		if err != nil {
			impl.logger.Errorw("error in getting k8s client for default cluster", "err", err)
			return ""
		}
	} else {
		_, k8sClient, err = impl.k8sCommonService.GetCoreClientByClusterId(authConfig.ClusterId)
		if err != nil {
			impl.logger.Errorw("error in getting k8s client for default cluster", "err", err)
			return ""
		}
	}
	devtronSecret, err := getSecret(authConfig.DevtronDexSecretNamespace, authConfig.DevtronSecretName, k8sClient)
	if err != nil {
		impl.logger.Errorw("error in getting devtron secret", "err", err)
	}
	secretData := devtronSecret.Data
	username, usernameOk := secretData[DEVTRON_ARGOCD_USERNAME_KEY]
	password, passwordOk := secretData[DEVTRON_ARGOCD_USER_PASSWORD_KEY]
	userNameStr := string(username)
	PasswordStr := string(password)
	if !usernameOk || !passwordOk {
		username, password, err := impl.createNewArgoCdUserForDevtron(k8sClient)
		if err != nil {
			impl.logger.Errorw("error in creating new argo cd user for devtron", "err", err)
		}
		userNameStr = username
		PasswordStr = password
	}
	isTokenAvailable := false
	for key, val := range secretData {
		if strings.HasPrefix(key, DEVTRON_ARGOCD_TOKEN_KEY) {
			isTokenAvailable = true
			token = string(val)
		}
	}
	if !isTokenAvailable {
		token, err = impl.createNewArgoCdTokenForDevtron(grpcConfig.ConnectionConfig, userNameStr, PasswordStr, 1, k8sClient)
		if err != nil {
			impl.logger.Errorw("error in creating new argo cd token for devtron", "err", err)
		}
	}
	return token
}

func (impl *ArgoCDConnectionManagerImpl) createNewArgoCdUserForDevtron(k8sClient *v1.CoreV1Client) (string, string, error) {
	username := DEVTRON_USER
	password := getNewPassword()
	userCapabilities := []string{ARGO_USER_APIKEY_CAPABILITY, ARGO_USER_LOGIN_CAPABILITY}
	//create new user at argo cd side
	err := impl.createNewArgoCdUser(username, password, userCapabilities, k8sClient)
	if err != nil {
		impl.logger.Errorw("error in creating new argocd user", "err", err)
		return "", "", err
	}
	//updating username and password in devtron-secret
	userCredentialMap := make(map[string]string)
	userCredentialMap[DEVTRON_ARGOCD_USERNAME_KEY] = username
	userCredentialMap[DEVTRON_ARGOCD_USER_PASSWORD_KEY] = password
	//updating username and password at devtron side
	err = impl.updateArgoCdUserInfoInDevtronSecret(userCredentialMap, k8sClient)
	if err != nil {
		impl.logger.Errorw("error in updating devtron-secret with argo-cd credentials", "err", err)
		return "", "", err
	}
	return username, password, nil
}

func (impl *ArgoCDConnectionManagerImpl) createNewArgoCdTokenForDevtron(connectionConfig *bean.Config, username, password string, tokenNo int, k8sClient *v1.CoreV1Client) (string, error) {
	//create new user at argo cd side
	token, err := impl.createTokenForArgoCdUser(connectionConfig, username, password)
	if err != nil {
		impl.logger.Errorw("error in creating new argocd user", "err", err)
		return "", err
	}
	//updating username and password in devtron-secret
	tokenMap := make(map[string]string)
	updatedTokenKey := fmt.Sprintf("%s_%d", DEVTRON_ARGOCD_TOKEN_KEY, tokenNo)
	tokenMap[updatedTokenKey] = token
	//updating username and password at devtron side
	err = impl.updateArgoCdUserInfoInDevtronSecret(tokenMap, k8sClient)
	if err != nil {
		impl.logger.Errorw("error in updating devtron-secret with argo-cd token", "err", err)
		return "", err
	}
	return token, nil
}

func (impl *ArgoCDConnectionManagerImpl) updateArgoCdUserInfoInDevtronSecret(userinfo map[string]string, k8sClient *v1.CoreV1Client) error {
	devtronSecret, err := getSecret(impl.devtronSecretConfig.DevtronDexSecretNamespace, impl.devtronSecretConfig.DevtronSecretName, k8sClient)
	if err != nil {
		impl.logger.Errorw("error in getting devtron secret", "err", err)
		return err
	}
	secretData := devtronSecret.Data
	if secretData == nil {
		secretData = make(map[string][]byte)
	}
	for key, value := range userinfo {
		secretData[key] = []byte(value)
	}
	devtronSecret.Data = secretData
	_, err = updateSecret(impl.devtronSecretConfig.DevtronDexSecretNamespace, devtronSecret, k8sClient)
	if err != nil {
		impl.logger.Errorw("error in updating devtron secret", "err", err)
		return err
	}
	return nil
}

func (impl *ArgoCDConnectionManagerImpl) createNewArgoCdUser(username, password string, capabilities []string, k8sClient *v1.CoreV1Client) error {
	//getting bcrypt hash of this password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		impl.logger.Errorw("error in getting bcrypt hash for password", "err", err)
		return err
	}
	//adding account name in configmap
	acdConfigmap, err := getConfigMap(DEVTRONCD_NAMESPACE, ARGOCD_CM, k8sClient)
	if err != nil {
		impl.logger.Errorw("error in getting argo cd configmap", "err", err)
		return err
	}
	cmData := acdConfigmap.Data
	if cmData == nil {
		cmData = make(map[string]string)
	}
	//updating data
	capabilitiesString := ""
	for i, capability := range capabilities {
		if i == 0 {
			capabilitiesString += capability
		} else {
			capabilitiesString += fmt.Sprintf(", %s", capability)
		}
	}
	newUserCmKey := fmt.Sprintf("accounts.%s", username)
	newUserCmValue := capabilitiesString
	cmData[newUserCmKey] = newUserCmValue
	acdConfigmap.Data = cmData
	_, err = updateConfigMap(DEVTRONCD_NAMESPACE, acdConfigmap, k8sClient)
	if err != nil {
		impl.logger.Errorw("error in updating argo cd configmap", "err", err)
		return err
	}
	acdSecret, err := getSecret(DEVTRONCD_NAMESPACE, ARGOCD_SECRET, k8sClient)
	if err != nil {
		impl.logger.Errorw("error in getting argo cd secret", "err", err)
		return err
	}
	secretData := acdSecret.Data
	if secretData == nil {
		secretData = make(map[string][]byte)
	}
	newUserSecretKey := fmt.Sprintf("accounts.%s.password", username)
	newUserSecretValue := passwordHash
	secretData[newUserSecretKey] = newUserSecretValue
	acdSecret.Data = secretData
	_, err = updateSecret(DEVTRONCD_NAMESPACE, acdSecret, k8sClient)
	if err != nil {
		impl.logger.Errorw("error in updating argo cd secret", "err", err)
		return err
	}
	return nil
}

func (impl *ArgoCDConnectionManagerImpl) createTokenForArgoCdUser(connectionConfig *bean.Config, username, password string) (string, error) {
	token, err := impl.passwordLogin(connectionConfig, username, password)
	if err != nil {
		impl.logger.Errorw("error in getting jwt token with username & password", "err", err)
		return "", err
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "token", token)
	clientConn := impl.getConnectionWithToken(connectionConfig, token)
	accountServiceClient := account.NewAccountServiceClient(clientConn)
	acdToken, err := accountServiceClient.CreateToken(ctx, &account.CreateTokenRequest{
		Name: username,
	})
	if err != nil {
		impl.logger.Errorw("error in creating acdToken in ArgoCd", "err", err)
		return "", err
	}

	// just checking and logging the ArgoCd version
	err = impl.versionService.CheckVersion(clientConn)
	if err != nil {
		impl.logger.Errorw("error found while checking ArgoCd Version", "err", err)
		return "", err
	}
	return acdToken.Token, nil
}

func (impl *ArgoCDConnectionManagerImpl) passwordLogin(connectionConfig *bean.Config, username, password string) (string, error) {
	conn := impl.getConnectionWithToken(connectionConfig, "")
	serviceClient := session.NewSessionServiceClient(conn)
	jwtToken, err := serviceClient.Create(context.Background(), username, password)
	return jwtToken, err
}

func getNewPassword() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, 16)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func getSecret(namespace string, name string, client *v1.CoreV1Client) (*apiv1.Secret, error) {
	secret, err := client.Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	} else {
		return secret, nil
	}
}

func updateSecret(namespace string, secret *apiv1.Secret, client *v1.CoreV1Client) (*apiv1.Secret, error) {
	secret, err := client.Secrets(namespace).Update(context.Background(), secret, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	} else {
		return secret, nil
	}
}

func getConfigMap(namespace string, name string, client *v1.CoreV1Client) (*apiv1.ConfigMap, error) {
	cm, err := client.ConfigMaps(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	} else {
		return cm, nil
	}
}

func updateConfigMap(namespace string, cm *apiv1.ConfigMap, client *v1.CoreV1Client) (*apiv1.ConfigMap, error) {
	cm, err := client.ConfigMaps(namespace).Update(context.Background(), cm, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	} else {
		return cm, nil
	}
}

func SettingsManager(cfg *bean.Config) (*settings.SettingsManager, error) {
	clientSet, kubeConfig := getK8sClient()
	namespace, _, err := kubeConfig.Namespace()
	if err != nil {
		return nil, err
	}
	//TODO: remove this hardcoding
	if len(cfg.Namespace) >= 0 {
		namespace = cfg.Namespace
	}
	return settings.NewSettingsManager(context.Background(), clientSet, namespace), nil
}

func getK8sClient() (k8sClient *kubernetes.Clientset, k8sConfig clientcmd.ClientConfig) {
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		panic(err)
	}
	clientSet := kubernetes.NewForConfigOrDie(config)
	return clientSet, kubeConfig
}

func (impl *ArgoCDConnectionManagerImpl) getArgoCdSettings() *settings.ArgoCDSettings {
	settings := impl.argoCDSettings
	if settings == nil {
		module, err := impl.moduleRepository.FindOne(ModuleNameArgoCd)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error on get acd connection", "err", err)
			return nil
		}
		if module == nil || module.Status != ModuleStatusInstalled {
			impl.logger.Errorw("error on get acd connection", "err", err)
			return nil
		}
		settings, err = impl.settingsManager.GetSettings()
		if err != nil {
			impl.logger.Errorw("error on get acd connection", "err", err)
			return nil
		}
		impl.argoCDSettings = settings
	}
	return impl.argoCDSettings
}
