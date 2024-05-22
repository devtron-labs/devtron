package bean

type ErrorResponse struct {
	Kind    string `json:"kind"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Reason  string `json:"reason"`
}
