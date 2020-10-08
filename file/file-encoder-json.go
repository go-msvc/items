package file

import (
	"encoding/json"
	"io"
	"os"

	"github.com/go-msvc/errors"
	"github.com/go-msvc/items"
)

type jsonFileEncoder struct {
	filename string
}

type fileItem struct {
	ID   string      `json:"id"`
	Data interface{} `json:"data"`
}

func (e jsonFileEncoder) Read() ([]items.IItem, error) {
	var f *os.File
	var err error
	if f, err = os.Open(e.filename); err != nil {
		if f, err = os.Create(e.filename); err != nil {
			return nil, errors.Errorf("cannot create/open %s", e.filename)
		}
	}
	defer f.Close()
	var fileItems []fileItem
	if err = json.NewDecoder(f).Decode(&fileItems); err != nil {
		if err != io.EOF {
			return nil, errors.Wrapf(err, "cannot read %s", e.filename)
		}
	}

	itemList := make([]items.IItem, len(fileItems))
	for index, fileItem := range fileItems {
		itemList[index] = items.NewItem(fileItem.ID, fileItem.Data)
	}
	return itemList, nil
}

func (e jsonFileEncoder) Write(itemList []items.IItem) error {
	var f *os.File
	var err error
	if f, err = os.Create(e.filename); err != nil {
		return errors.Errorf("cannot create/open %s", e.filename)
	}
	defer f.Close()

	fileItems := make([]fileItem, len(itemList))
	for index, item := range itemList {
		fileItems[index] = fileItem{ID: item.ID(), Data: item.Data()}
	}
	if err = json.NewEncoder(f).Encode(fileItems); err != nil {
		if err != io.EOF {
			return errors.Wrapf(err, "cannot write %s", e.filename)
		}
	}
	log.Debugf("wrote %d items to %s", len(fileItems), e.filename)
	return nil
}
