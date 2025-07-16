package pgvis

import (
	"encoding/json"
	"time"
)

type Feed struct {
	ID   int
	Time int64 // Time contains an UNIX millisecond timestamp
	Data any
}

func NewFeed(data any) *Feed {
	return &Feed{
		Time: time.Now().UnixMilli(),
		Data: data,
	}
}

func (f *Feed) DataToJSON() []byte {
	b, err := json.Marshal(f.Data)
	if err != nil {
		panic(err)
	}
	return b
}

func (f *Feed) JSONToData(b []byte) {
	err := json.Unmarshal(b, &f.Data)
	if err != nil {
		panic(err)
	}
}
