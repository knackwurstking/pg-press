package routes

type Global struct {
	Title            string
	SubTitle         string
	Version          string
	ServerPathPrefix string
}

func (g Global) HasSubTitle() bool {
	return g.SubTitle != ""
}
