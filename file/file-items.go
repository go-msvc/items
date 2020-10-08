package file

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/go-msvc/errors"
	"github.com/go-msvc/items"
)

//New() creates a new file item store
//Parameters:
//	filename      must exist or will be created
//	format        must be JSON or XML
//	tmplValue     is example value to restrict type of all values, nil to store any type
//	idGen         is item id generator, nil to use default
//	reload        channel to indicate when file changes must be loaded, nil not to reload ever
//                (if multiple writers, ids may duplicate, but reload is safe for one writer many readers)
func New(filename string, format string, tmplValue interface{}, idGen items.IIDGen, reload chan bool) (items.IItems, error) {
	f := fileItems{
		filename:     filename,
		itemDataType: reflect.TypeOf(tmplValue),
		idGen:        idGen,
		itemList:     []items.IItem{},
		itemByID:     map[string]items.IItem{},
		reload:       reload,
	}
	if f.idGen == nil {
		f.idGen = &incIDGen{last: 0}
	}
	//dereference pointer types
	for f.itemDataType != nil && f.itemDataType.Kind() == reflect.Ptr {
		f.itemDataType = f.itemDataType.Elem()
	}

	switch strings.ToUpper(format) {
	// case "XML":
	// 	f.encoder = xmlFileEncoder{filename: filename}
	case "JSON":
		f.encoder = jsonFileEncoder{filename: filename}
	default:
		return nil, errors.Errorf("unknown format:%s, expecting json|xml", format)
	}
	//read item list from file then populate id index
	if err := f.Reload(); err != nil {
		return nil, errors.Wrapf(err, "failed to load")
	}

	if f.reload != nil {
		go func(f *fileItems) {
			for <-f.reload {
				log.Debugf("Reloading...")
				if f.closed {
					log.Debugf("Reloading...")
					break //terminate when flag is set
				}
				if err := f.Reload(); err != nil {
					log.Debugf("Reloading...")
					log.Errorf("failed to reload: %v", err)
				}
				log.Debugf("Reloading...")
			}
			log.Debugf("Reloading...")
		}(&f)
	}
	return &f, nil
}

type IFileEncoder interface {
	Read() ([]items.IItem, error)
	Write([]items.IItem) error
}

type fileItems struct {
	sync.Mutex
	filename     string
	encoder      IFileEncoder
	itemDataType reflect.Type
	idGen        items.IIDGen
	reload       chan bool
	itemList     []items.IItem
	itemByID     map[string]items.IItem
	closed       bool
}

func (f *fileItems) Find(key map[string]interface{}) items.IItemSet {
	return nil
}

func (f *fileItems) Count() int {
	f.Lock()
	defer f.Unlock()
	return len(f.itemList)
}

func (f *fileItems) Add(data interface{}) (items.IItem, error) {
	f.Lock()
	defer f.Unlock()

	//check type
	if f.itemDataType != nil {
		t := reflect.TypeOf(data)
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		if t != f.itemDataType {
			return nil, errors.Errorf("cannot add %T to items %v", data, f.itemDataType)
		}
	}

	//todo: check for duplicates

	//todo: check indices

	newItem := items.NewItem(f.idGen.New(), data)

	if _, exists := f.itemByID[newItem.ID()]; exists {
		return nil, errors.Errorf("generator returns duplicate ID %s", newItem.ID())
	}
	f.itemList = append(f.itemList, newItem)
	f.itemByID[newItem.ID()] = newItem
	if err := f.encoder.Write(f.itemList); err != nil {
		return nil, errors.Wrapf(err, "failed to update file")
	}
	return newItem, nil
}

func (f *fileItems) Get(id string) items.IItem {
	f.Lock()
	defer f.Unlock()
	if foundItem, found := f.itemByID[id]; found {
		return foundItem
	}
	return nil
}

func (f *fileItems) Upd(id string, data interface{}) error {
	f.Lock()
	defer f.Unlock()
	return errors.Errorf("upd NYI")
}

func (f *fileItems) Del(id string) error {
	f.Lock()
	defer f.Unlock()
	return errors.Errorf("del NYI")
}

func (f *fileItems) Close() {
	f.closed = true
	if f.reload != nil {
		f.reload <- true
	}
}

func (f *fileItems) Reload() error {
	if f.closed {
		return errors.Errorf("not reloading closed file-items")
	}
	var err error
	if f.itemList, err = f.encoder.Read(); err != nil {
		return errors.Wrapf(err, "cannot read %s", f.filename)
	}
	for _, item := range f.itemList {
		f.itemByID[item.ID()] = item
		f.idGen.Used(item.ID())
	}
	log.Debugf("loaded %d=%d from %s", len(f.itemByID), len(f.itemList), f.filename)
	return nil
}

type incIDGen struct {
	sync.Mutex
	last int
}

func (ig *incIDGen) New() string {
	ig.Lock()
	defer ig.Unlock()
	ig.last++
	return fmt.Sprintf("%016x", ig.last)
}

func (ig *incIDGen) Used(id string) {
	if intID, err := strconv.Atoi(id); err == nil && intID > ig.last {
		ig.last = intID
	}
}
