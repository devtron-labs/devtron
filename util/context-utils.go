package util

import (
	"context"
	"fmt"
	"reflect"
)

func SetSuperAdminInContext(ctx context.Context, isSuperAdmin bool) context.Context {
	ctx = context.WithValue(ctx, "IsSuperAdmin", isSuperAdmin)
	return ctx
}

func GetIsSuperAdminFromContext(ctx context.Context) (bool, error) {
	flag := ctx.Value("isSuperAdmin")
	if reflect.TypeOf(flag).Kind() == reflect.Bool {
		return flag.(bool), nil
	}
	return false, fmt.Errorf("context not valid, isSuperAdmin not of type bool %v", flag)
}
