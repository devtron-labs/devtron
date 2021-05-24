package repository

import (
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/casbin"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"testing"
)

func TestUserAuthRepositoryImpl_generateDefaultPolicies(t *testing.T) {
	type fields struct {
		dbConnection *pg.DB
		Logger       *zap.SugaredLogger
	}
	type args struct {
		team       string
		entityName string
		env        string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []bean.PolicyRequest
		wantErr bool
	}{
		{
			name: "verfiy generated policy",
			fields: fields{Logger: dummyLogger()},
			args: args{
				team:       "dev",
				entityName: "applications",
				env:        "demo-devtron",
			},
			want: []bean.PolicyRequest{
				{
					Data: []casbin.Policy{
						{
							Type: "p",
							Sub:  "role:manager_dev_demo-devtron_applications",
							Res:  "applications",
							Act:  "*",
							Obj:  "dev/applications",
						},
						{
							Type: "p",
							Sub:  "role:manager_dev_demo-devtron_applications",
							Res:  "environment",
							Act:  "*",
							Obj:  "demo-devtron/applications",
						},
						{
							Type: "p",
							Sub:  "role:manager_dev_demo-devtron_applications",
							Res:  "team",
							Act:  "*",
							Obj:  "dev",
						},
						{
							Type: "p",
							Sub:  "role:manager_dev_demo-devtron_applications",
							Res:  "user",
							Act:  "*",
							Obj:  "dev",
						},
						{
							Type: "p",
							Sub:  "role:manager_dev_demo-devtron_applications",
							Res:  "notification",
							Act:  "*",
							Obj:  "dev",
						},
						{
							Type: "p",
							Sub:  "role:manager_dev_demo-devtron_applications",
							Res:  "global-environment",
							Act:  "*",
							Obj:  "demo-devtron",
						},
					},
				},
				{
					Data: []casbin.Policy{
						{
							Type: "p",
							Sub:  "role:admin_dev_demo-devtron_applications",
							Res:  "applications",
							Act:  "*",
							Obj:  "dev/applications",
						},
						{
							Type: "p",
							Sub:  "role:admin_dev_demo-devtron_applications",
							Res:  "environment",
							Act:  "*",
							Obj:  "demo-devtron/applications",
						},
						{
							Type: "p",
							Sub:  "role:admin_dev_demo-devtron_applications",
							Res:  "team",
							Act:  "get",
							Obj:  "dev",
						},
						{
							Type: "p",
							Sub:  "role:admin_dev_demo-devtron_applications",
							Res:  "global-environment",
							Act:  "get",
							Obj:  "demo-devtron",
						},
					},
				},
				{
					Data: []casbin.Policy{
						{
							Type: "p",
							Sub:  "role:trigger_dev_demo-devtron_applications",
							Res:  "applications",
							Act:  "get",
							Obj:  "dev/applications",
						},
						{
							Type: "p",
							Sub:  "role:trigger_dev_demo-devtron_applications",
							Res:  "applications",
							Act:  "trigger",
							Obj:  "dev/applications",
						},
						{
							Type: "p",
							Sub:  "role:trigger_dev_demo-devtron_applications",
							Res:  "environment",
							Act:  "trigger",
							Obj:  "demo-devtron/applications",
						},
						{
							Type: "p",
							Sub:  "role:trigger_dev_demo-devtron_applications",
							Res:  "environment",
							Act:  "get",
							Obj:  "demo-devtron/applications",
						},
						{
							Type: "p",
							Sub:  "role:trigger_dev_demo-devtron_applications",
							Res:  "global-environment",
							Act:  "get",
							Obj:  "demo-devtron",
						},
					},
				},
				{
					Data: []casbin.Policy{
						{
							Type: "p",
							Sub:  "role:view_dev_demo-devtron_applications",
							Res:  "applications",
							Act:  "get",
							Obj:  "dev/applications",
						},
						{
							Type: "p",
							Sub:  "role:view_dev_demo-devtron_applications",
							Res:  "environment",
							Act:  "get",
							Obj:  "demo-devtron/applications",
						},
						{
							Type: "p",
							Sub:  "role:view_dev_demo-devtron_applications",
							Res:  "global-environment",
							Act:  "get",
							Obj:  "demo-devtron",
						},
						{
							Type: "p",
							Sub:  "role:view_dev_demo-devtron_applications",
							Res:  "team",
							Act:  "get",
							Obj:  "dev",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := UserAuthRepositoryImpl{
				dbConnection: tt.fields.dbConnection,
				Logger:       tt.fields.Logger,
			}
			got, err := impl.generateDefaultPolicies(tt.args.team, tt.args.entityName, tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateDefaultPolicies() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i := 0; i < len(tt.want); i++ {
				for j := 0; j < len(tt.want[i].Data); j++ {
					if tt.want[i].Data[j].Type != got[i].Data[j].Type || tt.want[i].Data[j].Res != got[i].Data[j].Res ||
						tt.want[i].Data[j].Act != got[i].Data[j].Act || tt.want[i].Data[j].Obj != got[i].Data[j].Obj ||
						tt.want[i].Data[j].Sub != got[i].Data[j].Sub {
						t.Errorf("generateDefaultPolicies() got = %v, want %v at [%d][%d]", got[i].Data[j], tt.want[i].Data[j], i, j)
					}
				}
			}
		})
	}
}

func dummyLogger() *zap.SugaredLogger {

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.Level(5))
	l, err := config.Build()
	if err != nil {
		panic("failed to create the default logger: " + err.Error())
	}
	return l.Sugar()
}