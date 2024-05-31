/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package batch

//type AppAction interface {
//	Execute(build *v1.Workflow, props v1.InheritedProps) error
//}
//
//type AppActionImpl struct {
//	logger     *zap.SugaredLogger
//	appRepo    pc.AppRepository
//	teamRepo   team.TeamRepository
//	appService pipeline.DbPipelineOrchestrator
//}
//
//func NewAppAction(logger *zap.SugaredLogger,
//	appRepo pc.AppRepository, teamRepo team.TeamRepository, appService pipeline.DbPipelineOrchestrator) *AppActionImpl {
//	return &AppActionImpl{
//		appRepo:    appRepo,
//		teamRepo:   teamRepo,
//		appService: appService,
//		logger:     logger,
//	}
//}
//
//var appExecutor = []func(impl AppActionImpl, app *v1.App) error{executeAppCreate}
//
//func (impl AppActionImpl) Execute(app *v1.App, props v1.InheritedProps) error {
//	errs := make([]string, 0)
//	for _, f := range appExecutor {
//		errs = util.AppendErrorString(errs, f(impl, app))
//	}
//	return util.GetErrorOrNil(errs)
//}
//
//func executeAppCreate(impl AppActionImpl, app *v1.App) error {
//	if app.Operation != v1.Create {
//		return nil
//	}
//	if app.Destination.App == nil || len(*app.Destination.App) == 0 {
//		return fmt.Errorf("app name cannot be empty in build pipeline creation")
//	}
//	team, err := impl.teamRepo.FindByTeamName(app.Team)
//	if err != nil {
//		return err
//	}
//	//TODO: userId
//	appRequest := bean.CreateAppDTO{
//		Id:         0,
//		AppName:    *app.Destination.App,
//		UserId:     0,
//		Material:   nil,
//		TeamId:     team.Id,
//		TemplateId: 0,
//	}
//	appRes, err := impl.appService.CreateApp(&appRequest)
//	if err != nil {
//		return err
//	}
//	bean.UpdateMaterialDTO{
//		AppId:    appRes.Id,
//		Material: nil,
//		UserId:   0,
//	}
//
//	return nil
//}
