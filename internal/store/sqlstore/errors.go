package sqlstore

type SqlError struct {
	error
	sql string
}
