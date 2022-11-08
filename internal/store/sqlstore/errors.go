package sqlstore

import (
	"fmt"
	"github.com/jackc/pgconn"
	"strings"
)

// pgError checks err implements postgres error or not. If implements then returns error with postgres format or returns error
func pgError(format string, err error) error {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		err = fmt.Errorf(
			"SQL error: %s, Detail: %s, Where: %s, State: %s, Code: %s",
			pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.SQLState(), pgErr.Code,
		)
	}
	return fmt.Errorf(format, err)
}

// debugQuery ...
func debugQuery(q string) string {
	q = strings.ReplaceAll(q, "\t", "")
	q = strings.ReplaceAll(q, "\n", " ")
	// this need if anywhere in query used spaces instead of \t
	q = strings.ReplaceAll(q, "    ", "")
	q = strings.ReplaceAll(q, "; ", ";")
	return q
}
