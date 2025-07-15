package pgvis

import "encoding/json"

type Feed[T any] struct {
	ID      int
	Time    int64 // Time contains an UNIX millisecond timestamp
	Content string
	Data    T
}

func (f *Feed[T]) DataToJSON() []byte {
	b, err := json.Marshal(f.Data)
	if err != nil {
		panic(err)
	}
	return b
}

func (f *Feed[T]) JSONToData(b []byte) {
	err := json.Unmarshal(b, &f.Data)
	if err != nil {
		panic(err)
	}
}
