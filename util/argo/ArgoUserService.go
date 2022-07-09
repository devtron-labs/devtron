package argo

import (
	"context"
	"fmt"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/account"
	"github.com/argoproj/argo-cd/v2/util/settings"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/client/argocdServer/session"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"math/rand"
	"strconv"
	"strings"
)

const (
	DEVTRON_USER                     = "devtron"
	DEVTRONCD_NAMESPACE              = "devtroncd"
	ARGOCD_CM                        = "argocd-cm"
	ARGOCD_SECRET                    = "argocd-secret"
	DEVTRON_SECRET                   = "devtron-secret"
	ARGO_USER_APIKEY_CAPABILITY      = "apiKey"
	ARGO_USER_LOGIN_CAPABILITY       = "login"
	DEVTRON_ARGOCD_USERNAME_KEY      = "DEVTRON_ACD_USER_NAME"
	DEVTRON_ARGOCD_USER_PASSWORD_KEY = "DEVTRON_ACD_USER_PASSWORD"
	DEVTRON_ARGOCD_TOKEN_KEY         = "DEVTRON_ACD_TOKEN"
)

type ArgoUserService interface {
	GetLatestDevtronArgoCdUserToken() (string, error)
}
type ArgoUserServiceImpl struct {
	logger         *zap.SugaredLogger
	clusterService cluster.ClusterService
	acdSettings    *settings.ArgoCDSettings
}

func NewArgoUserServiceImpl(Logger *zap.SugaredLogger,
	clusterService cluster.ClusterService,
	acdSettings *settings.ArgoCDSettings) (*ArgoUserServiceImpl, error) {
	argoUserServiceImpl := &ArgoUserServiceImpl{
		logger:         Logger,
		clusterService: clusterService,
		acdSettings:    acdSettings,
	}
	go argoUserServiceImpl.UpdateArgoCdUserDetail()
	return argoUserServiceImpl, nil
}

func (impl *ArgoUserServiceImpl) UpdateArgoCdUserDetail() {
	cluster, err := impl.clusterService.FindOne(cluster.DefaultClusterName)
	if err != nil {
		impl.logger.Errorw("error in getting default cluster", "err", err)
	}
	clusterConfig, err := impl.clusterService.GetClusterConfig(cluster)
	if err != nil {
		impl.logger.Errorw("error in getting default cluster config", "err", err)
	}
	k8sClient, err := getClient(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting k8s client for default cluster", "err", err)
	}
	devtronSecret, err := getSecret(DEVTRONCD_NAMESPACE, DEVTRON_SECRET, k8sClient)
	if err != nil {
		impl.logger.Errorw("error in getting devtron secret", "err", err)
	}
	secretData := devtronSecret.Data
	username, usernameOk := secretData[DEVTRON_ARGOCD_USERNAME_KEY]
	password, passwordOk := secretData[DEVTRON_ARGOCD_USER_PASSWORD_KEY]
	userNameStr := string(username)
	PasswordStr := string(password)
	if !usernameOk || !passwordOk {
		username, password, err := impl.CreateNewArgoCdUserForDevtron(k8sClient)
		if err != nil {
			impl.logger.Errorw("error in creating new argo cd user for devtron", "err", err)
		}
		userNameStr = username
		PasswordStr = password
	}
	isTokenAvailable := false
	for key, _ := range secretData {
		if strings.HasPrefix(key, DEVTRON_ARGOCD_TOKEN_KEY) {
			isTokenAvailable = true
		}
	}
	if !isTokenAvailable {
		_, err = impl.CreateNewArgoCdTokenForDevtron(userNameStr, PasswordStr, 1, k8sClient)
		if err != nil {
			impl.logger.Errorw("error in creating new argo cd token for devtron", "err", err)
		}
	}
}

func (impl *ArgoUserServiceImpl) CreateNewArgoCdUserForDevtron(k8sClient *v1.CoreV1Client) (string, string, error) {
	username := DEVTRON_USER
	password := getNewPassword()
	userCapabilities := []string{ARGO_USER_APIKEY_CAPABILITY, ARGO_USER_LOGIN_CAPABILITY}
	//create new user at argo cd side
	err := impl.CreateNewArgoCdUser(username, password, userCapabilities, k8sClient)
	if err != nil {
		impl.logger.Errorw("error in creating new argocd user", "err", err)
		return "", "", err
	}
	//updating username and password in devtron-secret
	userCredentialMap := make(map[string]string)
	userCredentialMap[DEVTRON_ARGOCD_USERNAME_KEY] = username
	userCredentialMap[DEVTRON_ARGOCD_USER_PASSWORD_KEY] = password
	//updating username and password at devtron side
	err = impl.UpdateArgoCdUserInfoInDevtronSecret(userCredentialMap, k8sClient)
	if err != nil {
		impl.logger.Errorw("error in updating devtron-secret with argo-cd credentials", "err", err)
		return "", "", err
	}
	return username, password, nil
}

func (impl *ArgoUserServiceImpl) CreateNewArgoCdTokenForDevtron(username, password string, tokenNo int, k8sClient *v1.CoreV1Client) (string, error) {
	//create new user at argo cd side
	token, err := impl.CreateTokenForArgoCdUser(username, password)
	if err != nil {
		impl.logger.Errorw("error in creating new argocd user", "err", err)
		return "", err
	}
	//updating username and password in devtron-secret
	tokenMap := make(map[string]string)
	updatedTokenKey := fmt.Sprintf("%s_%d", DEVTRON_ARGOCD_TOKEN_KEY, tokenNo)
	tokenMap[updatedTokenKey] = token
	//updating username and password at devtron side
	err = impl.UpdateArgoCdUserInfoInDevtronSecret(tokenMap, k8sClient)
	if err != nil {
		impl.logger.Errorw("error in updating devtron-secret with argo-cd token", "err", err)
		return "", err
	}
	return token, nil
}
func (impl *ArgoUserServiceImpl) GetLatestDevtronArgoCdUserToken() (string, error) {
	cluster, err := impl.clusterService.FindOne(cluster.DefaultClusterName)
	if err != nil {
		impl.logger.Errorw("error in getting default cluster", "err", err)
		return "", err
	}
	clusterConfig, err := impl.clusterService.GetClusterConfig(cluster)
	if err != nil {
		impl.logger.Errorw("error in getting default cluster config", "err", err)
		return "", err
	}
	k8sClient, err := getClient(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting k8s client for default cluster", "err", err)
		return "", err
	}
	devtronSecret, err := getSecret(DEVTRONCD_NAMESPACE, DEVTRON_SECRET, k8sClient)
	if err != nil {
		impl.logger.Errorw("error in getting devtron secret", "err", err)
		return "", err
	}
	secretData := devtronSecret.Data
	username := secretData[DEVTRON_ARGOCD_USERNAME_KEY]
	password := secretData[DEVTRON_ARGOCD_USER_PASSWORD_KEY]
	latestTokenNo := 1
	isTokenAvailable := true
	var token string
	for key, value := range secretData {
		if strings.HasPrefix(key, DEVTRON_ARGOCD_TOKEN_KEY) {
			isTokenAvailable = true
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

	if !isTokenAvailable || len(token) == 0 {
		newTokenNo := latestTokenNo + 1
		token, err = impl.CreateNewArgoCdTokenForDevtron(string(username), string(password), newTokenNo, k8sClient)
		if err != nil {
			impl.logger.Errorw("error in creating new argo cd token for devtron", "err", err)
			return "", err
		}
	}
	return token, nil
}

func (impl *ArgoUserServiceImpl) UpdateArgoCdUserInfoInDevtronSecret(userinfo map[string]string, k8sClient *v1.CoreV1Client) error {
	devtronSecret, err := getSecret(DEVTRONCD_NAMESPACE, DEVTRON_SECRET, k8sClient)
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
	_, err = updateSecret(DEVTRONCD_NAMESPACE, devtronSecret, k8sClient)
	if err != nil {
		impl.logger.Errorw("error in updating devtron secret", "err", err)
		return err
	}
	return nil
}

func (impl *ArgoUserServiceImpl) CreateNewArgoCdUser(username, password string, capabilities []string, k8sClient *v1.CoreV1Client) error {
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

func (impl *ArgoUserServiceImpl) CreateTokenForArgoCdUser(username, password string) (string, error) {
	token, err := passwordLogin(impl.acdSettings, username, password)
	if err != nil {
		impl.logger.Errorw("error in getting jwt token with username & password", "err", err)
		return "", err
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "token", token)
	acdConnection := argocdServer.GetConnection(token, impl.acdSettings)
	accountServiceClient := account.NewAccountServiceClient(acdConnection)
	acdToken, err := accountServiceClient.CreateToken(ctx, &account.CreateTokenRequest{
		Name: username,
	})
	if err != nil {
		impl.logger.Errorw("error in creating token at argocd side", "err", err)
		return "", err
	}
	return acdToken.Token, nil
}

func passwordLogin(acdSettings *settings.ArgoCDSettings, username, password string) (string, error) {
	serviceClient := session.NewSessionServiceClient(acdSettings)
	token, err := serviceClient.Create(context.Background(), username, password)
	return token, err
}

func getNewPassword() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, 16)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func getClient(clusterConfig *util.ClusterConfig) (*v1.CoreV1Client, error) {
	cfg := &rest.Config{}
	cfg.Host = clusterConfig.Host
	cfg.BearerToken = clusterConfig.BearerToken
	cfg.Insecure = true
	client, err := v1.NewForConfig(cfg)
	return client, err
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
