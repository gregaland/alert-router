package routers

type Event struct {
	Id string
	Message string
}

type Router interface {
	Init() error
	GetConfig() interface{}
	Route(*Event, interface{}) error
}