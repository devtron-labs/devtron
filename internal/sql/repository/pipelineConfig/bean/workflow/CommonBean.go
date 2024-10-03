package workflow

type ArtifactUploadedType string

func (r ArtifactUploadedType) String() string {
	return string(r)
}

func GetArtifactUploadedType(isUploaded bool) ArtifactUploadedType {
	if isUploaded {
		return ArtifactUploaded
	}
	return ArtifactNotUploaded
}

func IsArtifactUploaded(s ArtifactUploadedType) (isArtifactUploaded bool, isMigrationRequired bool) {
	switch s {
	case ArtifactUploaded:
		return true, false
	case ArtifactNotUploaded:
		return false, false
	default:
		return false, true
	}
}

const (
	NullArtifactUploaded ArtifactUploadedType = "NA"
	ArtifactUploaded     ArtifactUploadedType = "Uploaded"
	ArtifactNotUploaded  ArtifactUploadedType = "NotUploaded"
)
