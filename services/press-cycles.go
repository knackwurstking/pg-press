package services

type PressCycles struct {
	*Base
}

func NewPressCycles(r *Registry) *PressCycles {
	return &PressCycles{
		Base: NewBase(r),
	}
}
