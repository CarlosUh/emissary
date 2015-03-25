package data

type DataSource interface {
	Next() (*Datum, error)
	HasNext() bool
}
