package bean

import "context"

type RequestContext struct {
	context.Context
}

func (context RequestContext) GetUser() UserContext {
	return context.Context.Value("user").(UserContext) //TODO KB: check this
}

// should return user context, ci pipeline context, cd pipeline context,  App context
// shall we handle load context as well ??

type UserContext struct {
	UserId  int32
	EmailId string
	Token   string
}
