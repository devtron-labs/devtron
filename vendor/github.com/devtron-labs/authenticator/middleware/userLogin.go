package middleware

import (
	"encoding/json"
	"flag"
	"fmt"
	passwordutil "github.com/devtron-labs/authenticator/password"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type LoginService struct {
	sessionManager *SessionManager
	kubeconfig     *rest.Config
}

type LocalDevMode bool

func NewUserLogin(sessionManager *SessionManager, devMode LocalDevMode) (*LoginService, error) {
	cfg, err := getKubeConfig(devMode)
	if err != nil {
		return nil, err
	}
	return &LoginService{
		sessionManager: sessionManager,
		kubeconfig:     cfg,
	}, nil
}
func (impl LoginService) CreateLoginSession(username string, password string) (string, error) {
	if username == "" || password == "" {
		return "", status.Errorf(codes.Unauthenticated, "no credentials supplied")
	}
	err := impl.verifyUsernamePassword(username, password)
	if err != nil {
		return "", err
	}
	uniqueId, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	impl.sessionManager.GetUserSessionDuration().Seconds()
	jwtToken, err := impl.sessionManager.Create(
		fmt.Sprintf("%s", username),
		int64(impl.sessionManager.GetUserSessionDuration().Seconds()),
		uniqueId.String())
	return jwtToken, err
}

// VerifyUsernamePassword verifies if a username/password combo is correct
func (impl LoginService) verifyUsernamePassword(username string, password string) error {
	if password == "" {
		return status.Errorf(codes.Unauthenticated, blankPasswordError)
	}
	// Enforce maximum length of username on local accounts
	if len(username) > maxUsernameLength {
		return status.Errorf(codes.InvalidArgument, usernameTooLongError, maxUsernameLength)
	}
	account, err := impl.GetAccount(username)
	if err != nil {
		if errStatus, ok := status.FromError(err); ok && errStatus.Code() == codes.NotFound {
			err = InvalidLoginErr
		}
		return err
	}
	valid, _ := passwordutil.VerifyPassword(password, account.PasswordHash)
	if !valid {
		return InvalidLoginErr
	}
	if !account.Enabled {
		return status.Errorf(codes.Unauthenticated, accountDisabled, username)
	}
	if !account.HasCapability(AccountCapabilityLogin) {
		return status.Errorf(codes.Unauthenticated, userDoesNotHaveCapability, username, AccountCapabilityLogin)
	}
	return nil
}

type Account struct {
	PasswordHash  string
	PasswordMtime *time.Time
	Enabled       bool
	Capabilities  []AccountCapability
	Tokens        []Token
}
type AccountCapability string

const (
	AccountCapabilityLogin = "login"
)

type Token struct {
	ID        string `json:"id"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp,omitempty"`
}

// FormatPasswordMtime return the formatted password modify time or empty string of password modify time is nil.
func (a *Account) FormatPasswordMtime() string {
	if a.PasswordMtime == nil {
		return ""
	}
	return a.PasswordMtime.Format(time.RFC3339)
}

// FormatCapabilities returns comma separate list of user capabilities.
func (a *Account) FormatCapabilities() string {
	var items []string
	for i := range a.Capabilities {
		items = append(items, string(a.Capabilities[i]))
	}
	return strings.Join(items, ",")
}

// TokenIndex return an index of a token with the given identifier or -1 if token not found.
func (a *Account) TokenIndex(id string) int {
	for i := range a.Tokens {
		if a.Tokens[i].ID == id {
			return i
		}
	}
	return -1
}

// HasCapability return true if the account has the specified capability.
func (a *Account) HasCapability(capability AccountCapability) bool {
	for _, c := range a.Capabilities {
		if c == capability {
			return true
		}
	}
	return false
}

func (impl LoginService) GetAccount(name string) (*Account, error) {
	if name != "admin" {
		return nil, fmt.Errorf("no account supported: %s", name)
	}
	secret, cm, err := impl.getArgoConfig(impl.kubeconfig)
	if err != nil {
		return nil, err
	}
	account, err := parseAdminAccount(secret, cm)
	if err != nil {
		return nil, err
	}
	return account, nil
}

const (
	settingAdminPasswordHashKey = "admin.password"
	// settingAdminPasswordMtimeKey designates the key for a root password mtime inside a Kubernetes secret.
	settingAdminPasswordMtimeKey = "admin.passwordMtime"
	settingAdminEnabledKey       = "admin.enabled"
	settingAdminTokensKey        = "admin.tokens"

	ArgoCDConfigMapName = "argocd-cm"
	ArgoCDSecretName    = "argocd-secret"
	ArgocdNamespaceName = "devtroncd"
)

func parseAdminAccount(secret *v1.Secret, cm *v1.ConfigMap) (*Account, error) {
	adminAccount := &Account{Enabled: true, Capabilities: []AccountCapability{AccountCapabilityLogin}}
	if adminPasswordHash, ok := secret.Data[settingAdminPasswordHashKey]; ok {
		adminAccount.PasswordHash = string(adminPasswordHash)
	}
	if adminPasswordMtimeBytes, ok := secret.Data[settingAdminPasswordMtimeKey]; ok {
		if mTime, err := time.Parse(time.RFC3339, string(adminPasswordMtimeBytes)); err == nil {
			adminAccount.PasswordMtime = &mTime
		}
	}

	adminAccount.Tokens = make([]Token, 0)
	if tokensStr, ok := secret.Data[settingAdminTokensKey]; ok && string(tokensStr) != "" {
		if err := json.Unmarshal(tokensStr, &adminAccount.Tokens); err != nil {
			return nil, err
		}
	}
	if enabledStr, ok := cm.Data[settingAdminEnabledKey]; ok {
		if enabled, err := strconv.ParseBool(enabledStr); err == nil {
			adminAccount.Enabled = enabled
		} else {
			log.Warnf("ConfigMap has invalid key %s: %v", settingAdminTokensKey, err)
		}
	}

	return adminAccount, nil
}

//TODO use it as generic function across system
func getKubeConfig(devMode LocalDevMode) (*rest.Config, error) {
	if devMode {
		usr, err := user.Current()
		if err != nil {
			return nil, err
		}
		kubeconfig := flag.String("kubeconfig-authenticator", filepath.Join(usr.HomeDir, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		flag.Parse()
		restConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			return nil, err
		}
		return restConfig, nil
	} else {
		restConfig, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		return restConfig, nil
	}
}

func (impl LoginService) getArgoConfig(config *rest.Config) (secret *v1.Secret, cm *v1.ConfigMap, err error) {
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	secret, err = clientSet.CoreV1().Secrets(ArgocdNamespaceName).Get(ArgoCDSecretName, v12.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	cm, err = clientSet.CoreV1().ConfigMaps(ArgocdNamespaceName).Get(ArgoCDConfigMapName, v12.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	return secret, cm, nil
}
