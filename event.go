package gigo

const (
	No = iota
	Debug
	Info
	Err
)

type Eventer interface {
	Emit(tag string, level int, message string)
}
