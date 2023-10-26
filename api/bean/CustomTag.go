package bean

type CustomTag struct {
	EntityKey            int    `json:"entityKey"`
	EntityValue          string `json:"entityValue"`
	TagPattern           string `json:"tagPattern"`
	AutoIncreasingNumber int    `json:"counterX"`
	Metadata             string `json:"metadata"`
}

type CustomTagErrorResponse struct {
	ConflictingArtifactPath string `json:"conflictingLink"`
	TagPattern              string `json:"tagPattern"`
	AutoIncreasingNumber    int    `json:"counterX"`
	Message                 string `json:"message"`
}
