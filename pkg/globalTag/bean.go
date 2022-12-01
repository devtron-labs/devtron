package globalTag

type GlobalTagDto struct {
	Id                     int    `json:"id,notnull"`
	Key                    string `json:"key,notnull"`
	Description            string `json:"description,notnull"`
	MandatoryProjectIdsCsv string `json:"mandatoryProjectIdsCsv"`
	CreatedOnInMs          int64  `json:"createdOnInMs,notnull"`
	UpdatedOnInMs          int64  `json:"updatedOnInMs"`
}

type GlobalTagDtoForProject struct {
	Key         string `json:"key,notnull"`
	IsMandatory bool   `json:"isMandatory,notnull"`
}

type CreateGlobalTagsRequest struct {
	Tags []*CreateGlobalTagDto `json:"tags,notnull"`
}

type CreateGlobalTagDto struct {
	Key                    string `json:"key,notnull"`
	Description            string `json:"description,notnull"`
	MandatoryProjectIdsCsv string `json:"mandatoryProjectIdsCsv"`
}

type DeleteGlobalTagsRequest struct {
	Ids []int `json:"ids,notnull"`
}

type UpdateGlobalTagsRequest struct {
	Tags []*UpdateGlobalTagDto `json:"tags,notnull"`
}

type UpdateGlobalTagDto struct {
	Id                     int    `json:"id,notnull"`
	MandatoryProjectIdsCsv string `json:"mandatoryProjectIdsCsv"`
}
