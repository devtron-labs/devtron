package util

import (
	"context"
	"fmt"
	"reflect"
)

const IsSuperAdminFlag = "isSuperAdmin"

func SetSuperAdminInContext(ctx context.Context, isSuperAdmin bool) context.Context {
	ctx = context.WithValue(ctx, IsSuperAdminFlag, isSuperAdmin)
	return ctx
}

func GetIsSuperAdminFromContext(ctx context.Context) (bool, error) {
	flag := ctx.Value(IsSuperAdminFlag)

	if flag != nil && reflect.TypeOf(flag).Kind() == reflect.Bool {
		return flag.(bool), nil
	}
	return false, fmt.Errorf("context not valid, isSuperAdmin flag not set correctly %v", flag)
}
