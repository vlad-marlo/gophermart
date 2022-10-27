package sqlstore

import (
	"fmt"
	"github.com/lib/pq"
	"strings"
)

type SqlError struct {
	error
	sql string
}

func sqlErr(format string, err error, sql string) SqlError {
	return SqlError{fmt.Errorf(format, pgError(err)), debugQuery(sql)}
}

// pgError checks err implements postgres error or not. If implements then returns error with postgres format or returns error
func pgError(err error) error {
	if pgErr, ok := err.(*pq.Error); ok {
		return fmt.Errorf(
			"SQL error: %s, Detail: %s, Where: %s, Code: %s, State: %s",
			pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState(),
		)
	}
	return err
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
