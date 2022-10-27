package sqlstore

import (
	"fmt"
	"github.com/lib/pq"
)

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
