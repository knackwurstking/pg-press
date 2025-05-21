package types

type MetalSheetTableCell[T string | int | float64 | SacmiThickness] struct {
	Value T

	valueType string
}

func NewMetalSheetTableCell_Int(value int) MetalSheetTableCell[int] {
	return MetalSheetTableCell[int]{
		valueType: "int",
		Value:     value,
	}
}

func NewMetalSheetTableCell_Float64(value float64) MetalSheetTableCell[float64] {
	return MetalSheetTableCell[float64]{
		valueType: "float64",
		Value:     value,
	}
}

func NewMetalSheetTableCell_SacmiThickness(value SacmiThickness) MetalSheetTableCell[SacmiThickness] {
	return MetalSheetTableCell[SacmiThickness]{
		valueType: "SacmiThickness",
		Value:     value,
	}
}

func (tc MetalSheetTableCell[T]) IsInt() bool {
	return tc.valueType == "int"
}

func (tc MetalSheetTableCell[T]) IsFloat64() bool {
	return tc.valueType == "float64"
}

func (tc MetalSheetTableCell[T]) IsSacmiThickness() bool {
	return tc.valueType == "SacmiThickness"
}
