package util

import (
	"context"
	"fmt"
	"reflect"
)

const IsSuperAdminFlag = "isSuperAdmin"
const Token = "token"
const UserId = "userId"

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

type RequestCtx struct {
	token  *string
	userId *int32
	context.Context
}

func NewRequestCtx(ctx context.Context) *RequestCtx {
	return &RequestCtx{
		Context: ctx,
	}
}

func (r *RequestCtx) GetToken() string {
	if r.token != nil {
		return *r.token
	}
	token := r.Context.Value(Token).(string)
	r.token = &token
	return token
}

func (r *RequestCtx) GetUserId() int32 {
	if r.userId != nil {
		return *r.userId
	}
	userId := r.Context.Value(UserId).(int32)
	r.userId = &userId
	return userId
}
