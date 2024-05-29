/*
 * Copyright (c) 2024. Devtron Inc.
 */

package resourceFilter

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"log"
	"testing"
)

func TestCELServiceImpl_EvaluateCELRequest(t *testing.T) {
	type fields struct {
		Logger *zap.SugaredLogger
	}
	type args struct {
		request CELRequest
	}
	cm := `
{
  "apiVersion": "v1",
  "kind": "ConfigMap",
  "metadata": {
    "name": "game-demo"
  },
  "data": {
    "player_initial_lives": "3",
    "ui_properties_file_name": "user-interface.properties",
    "game.properties": "enemy.types=aliens,monsters\nplayer.maximum-lives=5    \n",
    "user-interface.properties": "color.good=purple\ncolor.bad=yellow\nallow.textmode=true\n"
  }
} `
	uns := unstructured.Unstructured{}
	b := []byte(cm)
	err := uns.UnmarshalJSON(b)
	if err != nil {
		log.Panic(err)
	}
	log, err := util.NewSugardLogger()
	if err != nil {
		log.Panic(err)
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:   "basic test",
			fields: fields{Logger: log},
			args: args{request: CELRequest{
				Expression: "has(self.data.player_initial_lives)",
				ExpressionMetadata: ExpressionMetadata{
					Params: []ExpressionParam{
						{
							ParamName: "self",
							Value:     uns.UnstructuredContent(),
							Type:      ParamTypeObject,
						},
					},
				},
			},
			},
			want: true,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return false
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &CELServiceImpl{
				Logger: tt.fields.Logger,
			}
			got, err := impl.EvaluateCELRequest(tt.args.request)
			if !tt.wantErr(t, err, fmt.Sprintf("EvaluateCELRequest(%v)", tt.args.request)) {
				return
			}
			assert.Equalf(t, tt.want, got, "EvaluateCELRequest(%v)", tt.args.request)
		})
	}
}
