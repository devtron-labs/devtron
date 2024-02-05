package globalTag

type GlobalTagDto struct {
	Id                     int    `json:"id,notnull"`
	Key                    string `json:"key,notnull"`
	Description            string `json:"description,notnull"`
	MandatoryProjectIdsCsv string `json:"mandatoryProjectIdsCsv"`
	Propagate              bool   `json:"propagate"`
	CreatedOnInMs          int64  `json:"createdOnInMs,notnull"`
	UpdatedOnInMs          int64  `json:"updatedOnInMs"`
}

type GlobalTagDtoForProject struct {
	Key         string `json:"key,notnull"`
	Description string `json:"description,notnull"`
	IsMandatory bool   `json:"isMandatory,notnull"`
	Propagate   bool   `json:"propagate"`
}

type CreateGlobalTagsRequest struct {
	Tags []*CreateGlobalTagDto `json:"tags,notnull"`
}

type CreateGlobalTagDto struct {
	Key                    string `json:"key,notnull"`
	Description            string `json:"description,notnull"`
	MandatoryProjectIdsCsv string `json:"mandatoryProjectIdsCsv"`
	Propagate              bool   `json:"propagate"`
}

type DeleteGlobalTagsRequest struct {
	Ids []int `json:"ids,notnull"`
}

type UpdateGlobalTagsRequest struct {
	Tags []*UpdateGlobalTagDto `json:"tags,notnull"`
}

type UpdateGlobalTagDto struct {
	Id                     int    `json:"id,notnull"`
	Key                    string `json:"key,notnull"`
	Description            string `json:"description,notnull"`
	MandatoryProjectIdsCsv string `json:"mandatoryProjectIdsCsv"`
	Propagate              bool   `json:"propagate"`
}
