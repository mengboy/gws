package gws

type Msg interface {
	GetRouter() string
	GetData() interface{}
}

type Message struct {
	Path string      `json:"path"`
	Data interface{} `json:"data"`
}

func (m *Message) GetRouter() string {
	return m.Path
}

func (m *Message) GetData() interface{} {
	return m.Data
}
