package api


type Api struct {
	C string //for now nothing here
}

func NewApi() (self *Api) {
	self = &Api{
		C: "abcdef",
	}
	return
}

// serialisable info about META
type Info struct {
	*Config
}

func (i *Info) Infoo() interface{} {
	return i
}


func NewInfo(c *Config) *Info {
	return &Info{c}
}
