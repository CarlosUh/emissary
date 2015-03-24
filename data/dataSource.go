package data

type DataSource interface {
	Next() (DataMap, error)
	HasNext() bool
}
