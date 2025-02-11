package sqlbuilder

import (
	"database/sql"
	"database/sql/driver"
)

type ScannerValuer interface {
	sql.Scanner
	driver.Valuer
}
