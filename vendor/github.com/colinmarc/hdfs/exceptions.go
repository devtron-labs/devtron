package hdfs

import "os"

const (
	fileNotFoundException     = "java.io.FileNotFoundException"
	permissionDeniedException = "org.apache.hadoop.security.AccessControlException"
)

func interpretException(exception string, err error) error {
	switch exception {
	case fileNotFoundException:
		return os.ErrNotExist
	case permissionDeniedException:
		return os.ErrPermission
	default:
		return err
	}
}
