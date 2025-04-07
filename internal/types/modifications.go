package types

type ModificationData interface {
	MetalSheet | MetalSheetPress | MetalSheetData
}

type Modification[T ModificationData] struct {
	User
	Time int `json:"time"`
	Data T   `json:"data"`
}

func (mb *Modification[T]) GetData() *T {
	return &mb.Data
}

type Modifications[T ModificationData] []Modification[T]

func (m *Modifications[T]) Add(mb *Modification[T]) {
	*m = append(*m, *mb)
}
