package services

type Tools struct {
	*Base
}

func NewTools(r *Registry) *Tools {
	return &Tools{
		Base: NewBase(r),
	}
}
