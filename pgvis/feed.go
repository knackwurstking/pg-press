package pgvis

import "encoding/json"

type Feed[T any] struct {
	ID      int
	Time    int64 // Time contains an UNIX millisecond timestamp
	Content string
	Data    T
}

func (f *Feed[T]) DataToSQLBlob() []byte {
	b, err := json.Marshal(f)
	if err != nil {
		panic(err)
	}
	return b
}

func (f *Feed[T]) SQLBlobToData(b []byte) {
	err := json.Unmarshal(b, f)
	if err != nil {
		panic(err)
	}
}
