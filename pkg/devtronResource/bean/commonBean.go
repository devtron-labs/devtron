package bean

type FilterCriteriaDecoder struct {
	Resource DevtronResourceKind
	Type     FilterCriteriaIdentifier
	Value    string
	ValueInt int
}

type FilterCriteriaIdentifier string

const (
	Identifier FilterCriteriaIdentifier = "identifier"
	Id         FilterCriteriaIdentifier = "id"
)

type SearchCriteriaDecoder struct {
	SearchBy SearchPropertyBy
	Value    string
}

type SearchPropertyBy string

const (
	ArtifactTag SearchPropertyBy = "artifactTag"
	ImageTag    SearchPropertyBy = "imageTag"
)
