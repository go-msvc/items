package items_test

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/go-msvc/items/file"
)

func Test1(t *testing.T) {
	tmpFile, _ := ioutil.TempFile("", "test.json")
	fn := tmpFile.Name()

	i, err := file.New(fn, "json", nil, nil, file.Watcher(fn, file.Write))
	if err != nil {
		t.Fatalf("new failed: %+v", err)
	}

	if i.Count() != 0 {
		t.Fatalf("count %d not 0", i.Count())
	}

	i1, err := i.Add(float64(1))
	if err != nil {
		t.Fatalf("add failed: %+v", err)
	}
	if i1Value, ok := i1.Data().(float64); !ok {
		t.Fatalf(" cannot get i1 float64 value")
	} else {
		if i1Value != 1 {
			t.Fatalf("i1.data=%f != 1", i1Value)
		}
	}

	if i.Count() != 1 {
		t.Fatalf("count %d not 1", i.Count())
	}

	i11 := i.Get(i1.ID())
	if i11 == nil {
		t.Fatalf("failed to get i1.id=%s", i1.ID())
	}
	if i11.ID() != i1.ID() {
		t.Fatalf("got i11.id=%s != %s", i11.ID(), i1.ID())
	}

	j, err := file.New(fn, "json", interface{}(nil), nil, nil) //j is not watching
	if err != nil {
		t.Fatalf("new failed: %+v", err)
	}
	if j.Count() != 1 {
		t.Fatalf("count %d not 1", j.Count())
	}
	j11 := j.Get(i1.ID())
	if j11 == nil {
		t.Fatalf("failed to get i1.id=%s", i1.ID())
	}
	if j11.ID() != i1.ID() {
		t.Fatalf("got j11.id=%s != %s", j11.ID(), i1.ID())
	}
	if j11.Data() != i1.Data() {
		t.Fatalf("got j11.data=(%T)%v != (%T)%v", j11.Data(), j11.Data(), i1.Data(), i1.Data())
	}

	//add to j, i must detect change and reload (j does not detect changes, only one writer)
	if _, err := j.Add(2); err != nil {
		t.Fatalf("failed to add 2 to j: %v", err)
	}
	time.Sleep(time.Second)
	if i.Count() != 2 {
		t.Fatalf("after add to j, i.count=%d != 2", i.Count())
	}
}

func TestSameTypeItems(t *testing.T) {
	//create with item type - all items must be that type
	fn := "/tmp/items_test"
	os.Remove(fn)

	//type itemData int

	i, err := file.New(fn, "json", "", nil, nil)
	if err != nil {
		t.Fatalf("new failed: %+v", err)
	}

	if i.Count() != 0 {
		t.Fatalf("count %d not 0", i.Count())
	}

	_, err = i.Add("strings may be added")
	if err != nil {
		t.Fatalf("add failed: %+v", err)
	}
	s := "aaa"
	_, err = i.Add(s)
	if err != nil {
		t.Fatalf("add failed: %+v", err)
	}
	_, err = i.Add(&s)
	if err != nil {
		t.Fatalf("add failed: %+v", err)
	}

	_, err = i.Add(float64(1))
	if err == nil {
		t.Fatalf("add float succeeded: %+v", err)
	}
}
