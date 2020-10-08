package file

import (
	"github.com/go-msvc/errors"
	"github.com/go-msvc/items"
)

type xmlFileEncoder struct {
	filename string
}

func (e xmlFileEncoder) Read() (map[string]items.IItem, error) {
	return nil, errors.Errorf("NYI")
}
func (e xmlFileEncoder) Write(map[string]items.IItem) error {
	return errors.Errorf("NYI")
}
