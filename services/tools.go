package services

const (
	ToolQuerySelect = `id, position, format, type, code, regenerating, is_dead, press, binding`
)

type Tools struct {
	*Base
}

func NewTools(r *Registry) *Tools {
	return &Tools{
		Base: NewBase(r),
	}
}
