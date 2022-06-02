package deployment

import (
	"encoding/json"
	"fmt"
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
	"os"
	"path/filepath"
	"strings"
	"time"
)

type DeploymentConfigRestHandler interface {
	CreateChartFromFile(w http.ResponseWriter, r *http.Request)
	SaveChart(w http.ResponseWriter, r *http.Request)
	GetUploadedCharts(w http.ResponseWriter, r *http.Request)
}

type DeploymentConfigRestHandlerImpl struct {
	Logger             *zap.SugaredLogger
	userAuthService    user.UserService
	enforcer           casbin.Enforcer
	validator          *validator.Validate
	refChartDir        chartRepoRepository.RefChartDir
	chartService       pipeline.ChartService
	chartRefRepository chartRepoRepository.ChartRefRepository
}

type DeploymentChartInfo struct {
	ChartName    string `json:"chartName"`
	ChartVersion string `json:"chartVersion"`
	Description  string `json:"description"`
	FileId       string `json:"fileId"`
	Action       string `json:"action"`
	Message      string `json:"message"`
}

func NewDeploymentConfigRestHandlerImpl(Logger *zap.SugaredLogger, userAuthService user.UserService, enforcer casbin.Enforcer, validator *validator.Validate,
	refChartDir chartRepoRepository.RefChartDir, chartService pipeline.ChartService, chartRefRepository chartRepoRepository.ChartRefRepository) *DeploymentConfigRestHandlerImpl {
	return &DeploymentConfigRestHandlerImpl{
		Logger:             Logger,
		userAuthService:    userAuthService,
		enforcer:           enforcer,
		validator:          validator,
		refChartDir:        refChartDir,
		chartService:       chartService,
		chartRefRepository: chartRefRepository,
	}
}

func (handler *DeploymentConfigRestHandlerImpl) CreateChartFromFile(w http.ResponseWriter, r *http.Request) {
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

	if err := r.ParseForm(); err != nil {
		handler.Logger.Errorw("request err, Corrupted form data", "err", err, "payload", file)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	err = handler.chartService.ValidateUploadedFileFormat(fileHeader.Filename)
	if err != nil {
		handler.Logger.Errorw("request err, Unsupported format", "err", err, "payload", file)
		common.WriteJsonResp(w, errors.New("Unsupported format file is uploaded, please upload file with .tgz extension"), nil, http.StatusBadRequest)
		return
	}

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		handler.Logger.Errorw("request err, File parsing error", "err", err, "payload", file)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	chartInfo, err := handler.chartService.ExtractChartIfMissing(fileBytes, string(handler.refChartDir), "")

	if err != nil {
		if chartInfo != nil && chartInfo.TemporaryFolder != "" {
			err1 := os.RemoveAll(chartInfo.TemporaryFolder)
			if err1 != nil {
				handler.Logger.Errorw("error in deleting temp dir ", "err", err1)
			}
		}
		if err.Error() == "Chart exists already, try uploading another chart" {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		common.WriteJsonResp(w, fmt.Errorf(err.Error()), nil, http.StatusBadRequest)
		return
	}

	chartRefs := &chartRepoRepository.ChartRef{
		Name:             chartInfo.ChartName,
		Version:          chartInfo.ChartVersion,
		Location:         chartInfo.ChartLocation,
		Active:           true,
		Default:          false,
		ChartData:        fileBytes,
		ChartDescription: chartInfo.Description,
		UserUploaded:     true,
		AuditLog: sql.AuditLog{
			CreatedBy: userId,
			CreatedOn: time.Now(),
			UpdatedOn: time.Now(),
			UpdatedBy: userId,
		},
	}

	chartsJson, _ := json.Marshal(chartRefs)
	err = ioutil.WriteFile(filepath.Join(chartInfo.TemporaryFolder, "output.json"), chartsJson, 0644)
	if err != nil {
		common.WriteJsonResp(w, fmt.Errorf(err.Error()), nil, http.StatusInternalServerError)
		return
	}

	pathList := strings.Split(chartInfo.TemporaryFolder, "/")
	chartData := &DeploymentChartInfo{
		ChartName:    chartInfo.ChartName,
		ChartVersion: chartInfo.ChartVersion,
		Description:  chartInfo.Description,
		FileId:       pathList[len(pathList)-1],
		Message:      chartInfo.Message,
	}

	common.WriteJsonResp(w, err, chartData, http.StatusOK)
	return
}

func (handler *DeploymentConfigRestHandlerImpl) SaveChart(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionDelete, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	var request DeploymentChartInfo
	err = decoder.Decode(&request)
	if err != nil {
		handler.Logger.Errorw("decode err", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	location := filepath.Join(string(handler.refChartDir), request.FileId)
	if request.Action == "Save" {
		file, err := ioutil.ReadFile(filepath.Join(location, "output.json"))
		if err != nil {
			handler.Logger.Errorw("Error reading output.json", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		chartRefs := &chartRepoRepository.ChartRef{}
		err = json.Unmarshal(file, &chartRefs)
		if err != nil {
			handler.Logger.Errorw("unmarshall err", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		chartRefs.ChartDescription = request.Description
		err = handler.chartRefRepository.Save(chartRefs)
		if err != nil {
			handler.Logger.Errorw("error in saving Chart", "err", err)
			common.WriteJsonResp(w, err, "Chart couldn't be saved", http.StatusInternalServerError)
			return
		}
	}

	if location != "" {
		err = os.RemoveAll(location)
		if err != nil {
			handler.Logger.Errorw("error in deleting temp dir ", "err", err)
		}
	}

	common.WriteJsonResp(w, err, "Processed successfully", http.StatusOK)
	return

}

func (handler *DeploymentConfigRestHandlerImpl) GetUploadedCharts(w http.ResponseWriter, r *http.Request) {

	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionDelete, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	charts, err := handler.chartService.FetchChartInfoByFlag(true)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, charts, http.StatusOK)
	return
}
