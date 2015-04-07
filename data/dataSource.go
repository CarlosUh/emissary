package data

type DataSource interface {
	Next() (Getter, error)
	HasNext() bool
}

type NilDataSource struct{}

func (n *NilDataSource) Next() (Getter, error) {
	return &Datum{nil}, nil
}

func (n *NilDataSource) HasNext() bool {
	return false
}
