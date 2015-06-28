package gigo

type Emitter interface {
	Emit(msg interface{}) error
}

type Plugin interface {
	Start() error
	Stop() error
}

type Input interface {
	Plugin
}

type Output interface {
	Plugin
	Emitter
}
