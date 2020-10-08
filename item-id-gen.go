package items

type IIDGen interface {
	New() string
	Used(id string) //mark id as used - if necessary
}

type defaultIdGen struct{}

func (ig defaultIdGen) New() string {
	return ""
}
