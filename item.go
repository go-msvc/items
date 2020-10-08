package items

type IItem interface {
	ID() string
	Data() interface{}
	//Get(name string) interface{}
}

func NewItem(id string, data interface{}) IItem {
	return Item{id: id, data: data}
}

type Item struct {
	id   string
	data interface{}
}

func (i Item) ID() string { return i.id }

func (i Item) Data() interface{} { return i.data }
