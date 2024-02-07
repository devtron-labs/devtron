package deployment

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef/bean"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/chart"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/gorilla/mux"
	"github.com/juju/errors"
	"go.uber.org/zap"
)

type DeploymentConfigRestHandler interface {
	CreateChartFromFile(w http.ResponseWriter, r *http.Request)
	SaveChart(w http.ResponseWriter, r *http.Request)
	DownloadChart(w http.ResponseWriter, r *http.Request)
	GetUploadedCharts(w http.ResponseWriter, r *http.Request)
}

type DeploymentConfigRestHandlerImpl struct {
	Logger          *zap.SugaredLogger
	userAuthService user.UserService
	enforcer        casbin.Enforcer
	chartService    chart.ChartService
	chartRefService chartRef.ChartRefService
}

type DeploymentChartInfo struct {
	ChartName    string `json:"chartName"`
	ChartVersion string `json:"chartVersion"`
	Description  string `json:"description"`
	FileId       string `json:"fileId"`
	Action       string `json:"action"`
	Message      string `json:"message"`
}

func NewDeploymentConfigRestHandlerImpl(Logger *zap.SugaredLogger, userAuthService user.UserService, enforcer casbin.Enforcer,
	chartService chart.ChartService, chartRefService chartRef.ChartRefService) *DeploymentConfigRestHandlerImpl {
	return &DeploymentConfigRestHandlerImpl{
		Logger:          Logger,
		userAuthService: userAuthService,
		enforcer:        enforcer,
		chartService:    chartService,
		chartRefService: chartRefService,
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

	err = handler.chartRefService.ValidateCustomChartUploadedFileFormat(fileHeader.Filename)
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

	chartInfo, err := handler.chartRefService.ExtractChartIfMissing(fileBytes, bean.RefChartDirPath, "")

	if err != nil {
		if chartInfo != nil && chartInfo.TemporaryFolder != "" {
			err1 := os.RemoveAll(chartInfo.TemporaryFolder)
			if err1 != nil {
				handler.Logger.Errorw("error in deleting temp dir ", "err", err1)
			}
		}
		if err.Error() == bean.ChartAlreadyExistsInternalError || err.Error() == bean.ChartNameReservedInternalError {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		common.WriteJsonResp(w, fmt.Errorf(err.Error()), nil, http.StatusBadRequest)
		return
	}

	chartRefs := &bean.CustomChartRefDto{
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

	location := filepath.Join(bean.RefChartDirPath, request.FileId)
	if request.Action == "Save" {
		file, err := ioutil.ReadFile(filepath.Join(location, "output.json"))
		if err != nil {
			handler.Logger.Errorw("Error reading output.json", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		customChartRefDto := &bean.CustomChartRefDto{}
		err = json.Unmarshal(file, &customChartRefDto)
		if err != nil {
			handler.Logger.Errorw("unmarshall err", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		customChartRefDto.ChartDescription = request.Description
		err = handler.chartRefService.SaveCustomChart(customChartRefDto)
		if err != nil {
			handler.Logger.Errorw("error in saving Chart", "err", err, "request", customChartRefDto)
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

func (handler *DeploymentConfigRestHandlerImpl) DownloadChart(w http.ResponseWriter, r *http.Request) {
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

	vars := mux.Vars(r)
	chartRefId, err := strconv.Atoi(vars["chartRefId"])
	if err != nil {
		handler.Logger.Errorw("error in parsing chartRefId", "chartRefId", chartRefId, "err", err)
		common.WriteJsonResp(w, fmt.Errorf("error in parsing chartRefId : %s must be integer", chartRefId), nil, http.StatusBadRequest)
		return
	}
	manifestByteArr, err := handler.chartRefService.GetChartInBytes(chartRefId)
	if err != nil {
		handler.Logger.Errorw("error in converting chart to bytes", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteOctetStreamResp(w, r, manifestByteArr, "")
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

	charts, err := handler.chartRefService.FetchCustomChartsInfo()
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, charts, http.StatusOK)
	return
}
