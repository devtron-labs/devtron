package bean

type FilterCriteriaDecoder struct {
	Kind     DevtronResourceKind
	SubKind  DevtronResourceKind
	Type     FilterCriteriaIdentifier
	Value    string
	ValueInt int
}

type FilterCriteriaIdentifier string

const (
	Identifier FilterCriteriaIdentifier = "identifier"
	Id         FilterCriteriaIdentifier = "id"
)

func (i FilterCriteriaIdentifier) ToString() string {
	return string(i)
}

type SearchCriteriaDecoder struct {
	SearchBy SearchPropertyBy
	Value    string
}

type SearchPropertyBy string

const (
	ArtifactTag SearchPropertyBy = "artifactTag"
	ImageTag    SearchPropertyBy = "imageTag"
)
