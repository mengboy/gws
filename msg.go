package gws

type Msg interface {
	GetRouter() string
	GetData() interface{}
}

type Message struct {
	Router string
	Data   interface{}
}

func (m *Message) GetRouter() string {
	return m.Router
}

func (m *Message) GetData() interface{} {
	return m.Data
}
