package shared

type Image struct {
	Name string
	Data []byte
}

func NewImage(name string, data []byte) *Image {
	return &Image{
		Name: name,
		Data: data,
	}
}
