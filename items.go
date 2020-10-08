package items

type IItems interface {
	Find(key map[string]interface{}) IItemSet
	Count() int
	Add(data interface{}) (IItem, error)
	Get(id string) IItem
	Upd(id string, data interface{}) error
	Del(id string) error
	Close()
}

type IItemsInMemory interface {
	IItems
	Reload() error
}

type IItemSet interface {
	Next() IItem
}
