package data

type DataSource interface {
	Next() (Getter, error)
	HasNext() bool
}
