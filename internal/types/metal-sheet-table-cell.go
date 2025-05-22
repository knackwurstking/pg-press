package types

type MetalSheetTableCell[T string | int | float64 | SacmiThickness] struct {
	ValueType string
	Value     T
}

func NewMetalSheetTableCell_Int(value int) MetalSheetTableCell[int] {
	return MetalSheetTableCell[int]{
		ValueType: "int",
		Value:     value,
	}
}

func NewMetalSheetTableCell_Float64(value float64) MetalSheetTableCell[float64] {
	return MetalSheetTableCell[float64]{
		ValueType: "float64",
		Value:     value,
	}
}

func NewMetalSheetTableCell_SacmiThickness(value SacmiThickness) MetalSheetTableCell[SacmiThickness] {
	return MetalSheetTableCell[SacmiThickness]{
		ValueType: "SacmiThickness",
		Value:     value,
	}
}

func (tc MetalSheetTableCell[T]) IsInt() bool {
	return tc.ValueType == "int"
}

func (tc MetalSheetTableCell[T]) IsFloat64() bool {
	return tc.ValueType == "float64"
}

func (tc MetalSheetTableCell[T]) IsSacmiThickness() bool {
	return tc.ValueType == "SacmiThickness"
}
