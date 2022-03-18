package deployment

import (
	"github.com/devtron-labs/devtron/api/restHandler/common"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/juju/errors"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"io/ioutil"
	"net/http"
	"time"
)

type DeploymentRestHandler interface {
	CreateChartFromFile(w http.ResponseWriter, r *http.Request)
}

type DeploymentRestHandlerImpl struct {
	Logger             *zap.SugaredLogger
	userAuthService    user.UserService
	enforcer           casbin.Enforcer
	validator          *validator.Validate
	refChartDir        pipeline.RefChartDir
	chartService       pipeline.ChartService
	chartRefRepository chartRepoRepository.ChartRefRepository
}

func NewDeploymentRestHandlerImpl(Logger *zap.SugaredLogger, userAuthService user.UserService, enforcer casbin.Enforcer, validator *validator.Validate,
	refChartDir pipeline.RefChartDir, chartService pipeline.ChartService, chartRefRepository chartRepoRepository.ChartRefRepository) *DeploymentRestHandlerImpl {
	return &DeploymentRestHandlerImpl{
		Logger:             Logger,
		userAuthService:    userAuthService,
		enforcer:           enforcer,
		validator:          validator,
		refChartDir:        refChartDir,
		chartService:       chartService,
		chartRefRepository: chartRefRepository,
	}
}

func (handler *DeploymentRestHandlerImpl) CreateChartFromFile(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	file, fileHeader, err := r.FormFile("BinaryFile")
	if err != nil {
		handler.Logger.Errorw("request err, File parsing error", "err", err, "payload", file)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	//chartName := strings.Split(fileHeader.Filename, ".")
	if err := r.ParseForm(); err != nil {
		handler.Logger.Errorw("request err, Corrupted form data", "err", err, "payload", file)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	err = handler.chartService.ValidateFileUploaded(fileHeader.Filename)
	if err != nil {
		handler.Logger.Errorw("request err, Unsupported format", "err", err, "payload", file)
		common.WriteJsonResp(w, err, "Unsupported format file is uploaded, please upload file with .tar.gz extension", http.StatusBadRequest)
		return
	}

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		handler.Logger.Errorw("request err, File parsing error", "err", err, "payload", file)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
	}

	chartLocation, chartName, chartVersion, err := handler.chartService.ExtractChartIfMissing(fileBytes, string(handler.refChartDir), "")

	chartRefs := &chartRepoRepository.ChartRef{
		Name:      chartName,
		Version:   chartVersion,
		Location:  chartLocation,
		Active:    true,
		Default:   false,
		ChartData: fileBytes,
		AuditLog: sql.AuditLog{
			CreatedBy: userId,
			CreatedOn: time.Now(),
			UpdatedOn: time.Now(),
			UpdatedBy: userId,
		},
	}

	err = handler.chartRefRepository.Save(chartRefs)
	if err != nil {
		handler.Logger.Errorw("error in saving ConfigMap, CallbackConfigMap", "err", err)
		common.WriteJsonResp(w, err, "Chart couldn't be saved", http.StatusBadRequest)
		return
	}
	common.WriteJsonResp(w, err, "Chart Saved Successfully", http.StatusOK)
	return
}
