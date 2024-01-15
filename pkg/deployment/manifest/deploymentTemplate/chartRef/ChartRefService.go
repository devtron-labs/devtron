package chartRef

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/util"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef/adapter"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef/bean"
	util2 "github.com/devtron-labs/devtron/util"
	dirCopy "github.com/otiai10/copy"
	"go.uber.org/zap"
	"io/ioutil"
	"k8s.io/helm/pkg/chartutil"
	"os"
	"path"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strconv"
	"strings"
)

type ChartRefService interface {
	GetDefault() (*bean.ChartRefDto, error)
	GetAll() ([]*bean.ChartRefDto, error)
	GetAllChartMetadata() (map[string]bean.ChartRefMetaData, error)
	FindById(chartRefId int) (*bean.ChartRefDto, error)
	FindByVersionAndName(version, name string) (*bean.ChartRefDto, error)
	FetchInfoOfChartConfiguredInApp(appId int) (*bean.ChartRefDto, error)
	ChartRefAutocomplete() ([]*bean.ChartRefAutocompleteDto, error)

	SaveCustomChart(req *bean.CustomChartRefDto) error
	FetchCustomChartsInfo() ([]*bean.ChartDto, error)
	ValidateCustomChartUploadedFileFormat(fileName string) error

	GetSchemaAndReadmeForTemplateByChartRefId(chartRefId int) ([]byte, []byte, error)

	ChartRefIdsCompatible(oldChartRefId int, newChartRefId int) (bool, string, string)

	CheckChartExists(chartRefId int) error
	GetRefChart(chartRefId int) (string, string, error, string, string)
	ExtractChartIfMissing(chartData []byte, refChartDir string, location string) (*bean.ChartDataInfo, error)
	GetCustomChartInBytes(chartRefId int) ([]byte, error)
}

type ChartRefServiceImpl struct {
	logger               *zap.SugaredLogger
	chartRefRepository   chartRepoRepository.ChartRefRepository
	chartTemplateService util.ChartTemplateService
}

func NewChartRefServiceImpl(logger *zap.SugaredLogger,
	chartRefRepository chartRepoRepository.ChartRefRepository,
	chartTemplateService util.ChartTemplateService) *ChartRefServiceImpl {
	// cache devtron reference charts list
	devtronChartList, _ := chartRefRepository.FetchAllChartInfoByUploadFlag(false)
	setReservedChartList(devtronChartList)
	return &ChartRefServiceImpl{
		logger:               logger,
		chartRefRepository:   chartRefRepository,
		chartTemplateService: chartTemplateService,
	}
}

func (impl *ChartRefServiceImpl) ValidateCustomChartUploadedFileFormat(fileName string) error {
	if !strings.HasSuffix(fileName, ".tgz") {
		return errors.New("unsupported format")
	}
	return nil
}

func (impl *ChartRefServiceImpl) GetDefault() (*bean.ChartRefDto, error) {
	chartRef, err := impl.chartRefRepository.GetDefault()
	if err != nil {
		impl.logger.Errorw("error in getting default chartRef", "err", err)
		return nil, err
	}
	return adapter.ConvertChartRefDbObjToBean(chartRef), nil
}

func (impl *ChartRefServiceImpl) GetAll() ([]*bean.ChartRefDto, error) {
	chartRefs, err := impl.chartRefRepository.GetAll()
	if err != nil {
		impl.logger.Errorw("error in getting all chartRefs", "err", err)
		return nil, err
	}
	chartRefDtos := make([]*bean.ChartRefDto, 0, len(chartRefs))
	for _, chartRef := range chartRefs {
		chartRefDtos = append(chartRefDtos, adapter.ConvertChartRefDbObjToBean(chartRef))
	}
	return chartRefDtos, nil
}

func (impl *ChartRefServiceImpl) GetAllChartMetadata() (map[string]bean.ChartRefMetaData, error) {
	chartRefMetadatas, err := impl.chartRefRepository.GetAllChartMetadata()
	if err != nil {
		impl.logger.Errorw("error in getting all chart metadatas", "err", err)
		return nil, err
	}
	chartsMetadataMap := make(map[string]bean.ChartRefMetaData, len(chartRefMetadatas))
	for _, chartRefMetadata := range chartRefMetadatas {
		metadataDto := bean.ChartRefMetaData{
			ChartDescription: chartRefMetadata.ChartDescription,
		}
		chartsMetadataMap[chartRefMetadata.ChartName] = metadataDto
	}
	return chartsMetadataMap, nil
}

func (impl *ChartRefServiceImpl) ChartRefIdsCompatible(oldChartRefId int, newChartRefId int) (bool, string, string) {
	oldChart, err := impl.FindById(oldChartRefId)
	if err != nil {
		return false, "", ""
	}
	newChart, err := impl.FindById(newChartRefId)
	if err != nil {
		return false, "", ""
	}
	return CheckCompatibility(oldChart.Name, newChart.Name), oldChart.Name, newChart.Name
}

func (impl *ChartRefServiceImpl) FindById(chartRefId int) (*bean.ChartRefDto, error) {
	chartRef, err := impl.chartRefRepository.FindById(chartRefId)
	if err != nil {
		impl.logger.Errorw("error in getting chartRef by id", "err", err, "chartRefId", chartRefId)
		return nil, err
	}
	return adapter.ConvertChartRefDbObjToBean(chartRef), nil
}

func (impl *ChartRefServiceImpl) FindByVersionAndName(version, name string) (*bean.ChartRefDto, error) {
	chartRef, err := impl.chartRefRepository.FindByVersionAndName(name, version)
	if err != nil {
		impl.logger.Errorw("error in getting chartRef by version and name", "err", err, "version", version, "name", name)
		return nil, err
	}
	return adapter.ConvertChartRefDbObjToBean(chartRef), nil
}

func (impl *ChartRefServiceImpl) FetchInfoOfChartConfiguredInApp(appId int) (*bean.ChartRefDto, error) {
	chartRef, err := impl.chartRefRepository.FetchInfoOfChartConfiguredInApp(appId)
	if err != nil {
		impl.logger.Errorw("error in getting chart info for chart configured in app", "err", err, "appId", appId)
		return nil, err
	}
	return adapter.ConvertChartRefDbObjToBean(chartRef), nil
}

func (impl *ChartRefServiceImpl) SaveCustomChart(req *bean.CustomChartRefDto) error {
	chartRefDbObj := adapter.ConvertCustomChartRefDtoToDbObj(req)
	err := impl.chartRefRepository.Save(chartRefDbObj)
	if err != nil {
		impl.logger.Errorw("error in saving chart ref", "err", err, "chartRef", chartRefDbObj)
		return err
	}
	return nil
}

func (impl *ChartRefServiceImpl) GetRefChart(chartRefId int) (string, string, error, string, string) {
	var template string
	var version string
	//path of file in chart from where strategy config is to be taken
	var pipelineStrategyPath string
	if chartRefId > 0 {
		chartRef, err := impl.chartRefRepository.FindById(chartRefId)
		if err != nil {
			chartRef, err = impl.chartRefRepository.GetDefault()
			if err != nil {
				return "", "", err, "", ""
			}
		} else if chartRef.UserUploaded {
			refChartLocation := filepath.Join(chartRepoRepository.RefChartDirPath, chartRef.Location)
			if _, err := os.Stat(refChartLocation); os.IsNotExist(err) {
				chartInfo, err := impl.ExtractChartIfMissing(chartRef.ChartData, chartRepoRepository.RefChartDirPath, chartRef.Location)
				if chartInfo != nil && chartInfo.TemporaryFolder != "" {
					err1 := os.RemoveAll(chartInfo.TemporaryFolder)
					if err1 != nil {
						impl.logger.Errorw("error in deleting temp dir ", "err", err)
					}
				}
				if err != nil {
					impl.logger.Errorw("Error regarding uploaded chart", "err", err)
					return "", "", err, "", ""
				}

			}
		}
		template = chartRef.Location
		version = chartRef.Version
		pipelineStrategyPath = chartRef.DeploymentStrategyPath
	} else {
		chartRef, err := impl.chartRefRepository.GetDefault()
		if err != nil {
			return "", "", err, "", ""
		}
		template = chartRef.Location
		version = chartRef.Version
		pipelineStrategyPath = chartRef.DeploymentStrategyPath
	}

	//TODO VIKI- fetch from chart ref table
	chartPath := path.Join(chartRepoRepository.RefChartDirPath, template)
	valid, err := chartutil.IsChartDir(chartPath)
	if err != nil || !valid {
		impl.logger.Errorw("invalid base chart", "dir", chartPath, "err", err)
		return "", "", err, "", ""
	}
	return chartPath, template, nil, version, pipelineStrategyPath
}

func (impl *ChartRefServiceImpl) GetSchemaAndReadmeForTemplateByChartRefId(chartRefId int) ([]byte, []byte, error) {
	refChart, _, err, _, _ := impl.GetRefChart(chartRefId)
	if err != nil {
		impl.logger.Errorw("error in getting refChart", "err", err, "chartRefId", chartRefId)
		return nil, nil, err
	}
	var schemaByte []byte
	var readmeByte []byte
	err = impl.CheckChartExists(chartRefId)
	if err != nil {
		impl.logger.Errorw("error in getting refChart", "err", err, "chartRefId", chartRefId)
		return nil, nil, err
	}
	schemaByte, err = ioutil.ReadFile(filepath.Clean(filepath.Join(refChart, "schema.json")))
	if err != nil {
		impl.logger.Errorw("error in reading schema.json file for refChart", "err", err, "chartRefId", chartRefId)
	}
	readmeByte, err = ioutil.ReadFile(filepath.Clean(filepath.Join(refChart, "README.md")))
	if err != nil {
		impl.logger.Errorw("error in reading readme file for refChart", "err", err, "chartRefId", chartRefId)
	}
	return schemaByte, readmeByte, nil
}

func (impl *ChartRefServiceImpl) ChartRefAutocomplete() ([]*bean.ChartRefAutocompleteDto, error) {
	var chartRefs []*bean.ChartRefAutocompleteDto
	results, err := impl.chartRefRepository.GetAll()
	if err != nil {
		impl.logger.Errorw("error in fetching chart config", "err", err)
		return chartRefs, err
	}

	for _, result := range results {
		chartRefs = append(chartRefs, &bean.ChartRefAutocompleteDto{
			Id:                    result.Id,
			Version:               result.Version,
			Description:           result.ChartDescription,
			UserUploaded:          result.UserUploaded,
			IsAppMetricsSupported: result.IsAppMetricsSupported,
		})
	}

	return chartRefs, nil
}

func (impl *ChartRefServiceImpl) GetCustomChartInBytes(chartRefId int) ([]byte, error) {
	chartRef, err := impl.chartRefRepository.FindById(chartRefId)
	if err != nil {
		impl.logger.Errorw("error getting chart data", "chartRefId", chartRefId, "err", err)
		return nil, err
	}
	// For user uploaded charts ChartData will be retrieved from DB
	if chartRef.ChartData != nil {
		return chartRef.ChartData, nil
	}
	// For Devtron reference charts the chart will be load from the directory location
	refChartPath := filepath.Join(chartRepoRepository.RefChartDirPath, chartRef.Location)
	manifestByteArr, err := impl.chartTemplateService.LoadChartInBytes(refChartPath, false)
	if err != nil {
		impl.logger.Errorw("error in converting chart to bytes", "err", err)
		return nil, err
	}
	return manifestByteArr, nil
}

func (impl *ChartRefServiceImpl) FetchCustomChartsInfo() ([]*bean.ChartDto, error) {
	resultsMetadata, err := impl.chartRefRepository.GetAllChartMetadata()
	if err != nil {
		impl.logger.Errorw("error in fetching chart metadata", "err", err)
		return nil, err
	}
	chartsMetadata := make(map[string]string)
	for _, resultMetadata := range resultsMetadata {
		chartsMetadata[resultMetadata.ChartName] = resultMetadata.ChartDescription
	}
	repo, err := impl.chartRefRepository.GetAll()
	if err != nil {
		return nil, err
	}
	var chartDtos []*bean.ChartDto
	for _, ref := range repo {
		if len(ref.Name) == 0 {
			ref.Name = bean.RolloutChartType
		}
		if description, ok := chartsMetadata[ref.Name]; ref.ChartDescription == "" && ok {
			ref.ChartDescription = description
		}
		chartDto := &bean.ChartDto{
			Id:               ref.Id,
			Name:             ref.Name,
			ChartDescription: ref.ChartDescription,
			Version:          ref.Version,
			IsUserUploaded:   ref.UserUploaded,
		}
		chartDtos = append(chartDtos, chartDto)
	}
	return chartDtos, err
}

func (impl *ChartRefServiceImpl) CheckChartExists(chartRefId int) error {
	chartRefValue, err := impl.chartRefRepository.FindById(chartRefId)
	if err != nil {
		impl.logger.Errorw("error in finding ref chart by id", "err", err)
		return err
	}
	refChartLocation := filepath.Join(chartRepoRepository.RefChartDirPath, chartRefValue.Location)
	if _, err := os.Stat(refChartLocation); os.IsNotExist(err) {
		chartInfo, err := impl.ExtractChartIfMissing(chartRefValue.ChartData, chartRepoRepository.RefChartDirPath, chartRefValue.Location)
		if chartInfo != nil && chartInfo.TemporaryFolder != "" {
			err1 := os.RemoveAll(chartInfo.TemporaryFolder)
			if err1 != nil {
				impl.logger.Errorw("error in deleting temp dir ", "err", err)
			}
		}
		return err
	}
	return nil
}

func (impl *ChartRefServiceImpl) ExtractChartIfMissing(chartData []byte, refChartDir string, location string) (*bean.ChartDataInfo, error) {
	binaryDataReader := bytes.NewReader(chartData)
	dir := impl.chartTemplateService.GetDir()
	chartInfo := &bean.ChartDataInfo{
		ChartName:       "",
		ChartVersion:    "",
		ChartLocation:   "",
		TemporaryFolder: "",
		Description:     "",
		Message:         "",
	}
	temporaryChartWorkingDir := filepath.Clean(filepath.Join(refChartDir, dir))
	err := os.MkdirAll(temporaryChartWorkingDir, os.ModePerm)
	if err != nil {
		impl.logger.Errorw("error in creating directory, CallbackConfigMap", "err", err)
		return chartInfo, err
	}
	chartInfo.TemporaryFolder = temporaryChartWorkingDir
	err = util2.ExtractTarGz(binaryDataReader, temporaryChartWorkingDir)
	if err != nil {
		impl.logger.Errorw("error in extracting binary data of charts", "err", err)
		return chartInfo, err
	}

	var chartLocation string
	var chartName string
	var chartVersion string
	var fileName string

	files, err := ioutil.ReadDir(temporaryChartWorkingDir)
	if err != nil {
		impl.logger.Errorw("error in reading err dir", "err", err)
		return chartInfo, err
	}

	fileName = files[0].Name()
	if strings.HasPrefix(files[0].Name(), ".") {
		fileName = files[1].Name()
	}

	currentChartWorkingDir := filepath.Clean(filepath.Join(temporaryChartWorkingDir, fileName))

	if location == "" {
		chartYaml, err := impl.readChartMetaDataForLocation(temporaryChartWorkingDir, fileName)
		var errorList error
		if err != nil {
			impl.logger.Errorw("Chart yaml file or content not found")
			errorList = err
		}

		err = util2.CheckForMissingFiles(currentChartWorkingDir)
		if err != nil {
			impl.logger.Errorw("Missing files in the folder", "err", err)
			if errorList != nil {
				errorList = errors.New(errorList.Error() + "; " + err.Error())
			} else {
				errorList = err
			}

		}

		if errorList != nil {
			return chartInfo, errorList
		}

		chartName = chartYaml.Name
		chartVersion = chartYaml.Version
		chartInfo.Description = chartYaml.Description
		chartLocation = impl.getLocationFromChartNameAndVersion(chartName, chartVersion)
		location = chartLocation

		// Validate: chart name shouldn't conflict with Devtron charts (no user uploaded chart names should contain any devtron chart names as the prefix)
		isReservedChart, _ := impl.validateReservedChartName(chartName)
		if isReservedChart {
			impl.logger.Errorw("request err, chart name is reserved by Devtron")
			err = &util.ApiError{
				Code:            constants.ChartNameAlreadyReserved,
				InternalMessage: bean.CHART_NAME_RESERVED_INTERNAL_ERROR,
				UserMessage:     fmt.Sprintf("The name '%s' is reserved for a chart provided by Devtron", chartName),
			}
			return chartInfo, err
		}

		// Validate: chart location should be unique
		exists, err := impl.chartRefRepository.CheckIfDataExists(location)
		if err != nil {
			impl.logger.Errorw("Error in searching the database")
			return chartInfo, err
		}
		if exists {
			impl.logger.Errorw("request err, chart name and version exists already in the database")
			err = &util.ApiError{
				Code:            constants.ChartCreatedAlreadyExists,
				InternalMessage: bean.CHART_ALREADY_EXISTS_INTERNAL_ERROR,
				UserMessage:     fmt.Sprintf("%s of %s exists already in the database", chartVersion, chartName),
			}
			return chartInfo, err
		}

		//User Info Message: uploading new version of the existing chart name
		existingChart, err := impl.chartRefRepository.FetchChart(chartName)
		if err == nil && existingChart != nil {
			chartInfo.Message = "New Version detected for " + existingChart[0].Name
		}

	} else {
		err = dirCopy.Copy(currentChartWorkingDir, filepath.Clean(filepath.Join(refChartDir, location)))
		if err != nil {
			impl.logger.Errorw("error in copying chart from temp dir to ref chart dir", "err", err)
			return chartInfo, err
		}
	}

	chartInfo.ChartLocation = location
	chartInfo.ChartName = chartName
	chartInfo.ChartVersion = chartVersion
	return chartInfo, nil
}

func (impl *ChartRefServiceImpl) readChartMetaDataForLocation(chartDir string, fileName string) (*bean.ChartYamlStruct, error) {
	chartLocation := filepath.Clean(filepath.Join(chartDir, fileName))

	chartYamlPath := filepath.Clean(filepath.Join(chartLocation, "Chart.yaml"))
	if _, err := os.Stat(chartYamlPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Chart.yaml file not present in the directory")
	}

	data, err := ioutil.ReadFile(chartYamlPath)
	if err != nil {
		impl.logger.Errorw("failed reading data from file", "err", err)
		return nil, err
	}
	//println(data)
	var chartYaml bean.ChartYamlStruct
	err = yaml.Unmarshal(data, &chartYaml)
	if err != nil {
		impl.logger.Errorw("Unmarshal error of yaml file", "err", err)
		return nil, err
	}
	if chartYaml.Name == "" || chartYaml.Version == "" {
		impl.logger.Errorw("Missing values in yaml file either name or version", "err", err)
		return nil, errors.New("Missing values in yaml file either name or version")
	}
	ver := strings.Split(chartYaml.Version, ".")
	if len(ver) == 3 {
		for _, verObject := range ver {
			if _, err := strconv.ParseInt(verObject, 10, 64); err != nil {
				return nil, errors.New("Version should contain integers (Ex: 1.1.0)")
			}
		}
		return &chartYaml, nil
	}
	return nil, errors.New("Version should be of length 3 integers with dot seperated (Ex: 1.1.0)")
}

func (impl *ChartRefServiceImpl) validateReservedChartName(chartName string) (isReservedChart bool, err error) {
	formattedChartName := impl.formatChartName(chartName)
	for _, reservedChart := range *bean.ReservedChartRefNamesList {
		isReservedChart = (reservedChart.LocationPrefix != "" && strings.HasPrefix(formattedChartName, reservedChart.LocationPrefix)) ||
			(reservedChart.Name != "" && strings.HasPrefix(strings.ToLower(chartName), reservedChart.Name))
		if isReservedChart {
			return true, nil
		}
	}
	return false, nil
}

func (impl *ChartRefServiceImpl) getLocationFromChartNameAndVersion(chartName string, chartVersion string) string {
	var chartLocation string
	chartname := impl.formatChartName(chartName)
	chartversion := strings.ReplaceAll(chartVersion, ".", "-")
	if !strings.Contains(chartname, chartversion) {
		chartLocation = chartname + "_" + chartversion
	} else {
		chartLocation = chartname
	}
	return chartLocation
}

func (impl *ChartRefServiceImpl) formatChartName(chartName string) string {
	chartname := strings.ReplaceAll(chartName, ".", "-")
	chartname = strings.ReplaceAll(chartname, " ", "_")
	return chartname
}

func setReservedChartList(devtronChartList []*chartRepoRepository.ChartRef) {
	reservedChartRefNamesList := []bean.ReservedChartList{
		{Name: strings.ToLower(bean.RolloutChartType), LocationPrefix: ""},
		{Name: "", LocationPrefix: bean.ReferenceChart},
	}
	for _, devtronChart := range devtronChartList {
		reservedChartRefNamesList = append(reservedChartRefNamesList, bean.ReservedChartList{
			Name:           strings.ToLower(devtronChart.Name),
			LocationPrefix: strings.Split(devtronChart.Location, "_")[0],
		})
	}
	bean.ReservedChartRefNamesList = &reservedChartRefNamesList
}
