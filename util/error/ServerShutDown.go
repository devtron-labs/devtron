package error

// ServerShutDown is the error returned by Context when the context
// is canceled by server stop signal
var ServerShutDown error = serverShutDownError{}

type serverShutDownError struct{}

func (serverShutDownError) Error() string { return "server stopped: context canceled" }
