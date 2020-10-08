package file_test

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/go-msvc/items/file"
)

func TestWatcherWrite(t *testing.T) {
	tmpFile, err := ioutil.TempFile("/tmp", "test.json")
	if err != nil {
		t.Fatalf("tmpfile: %v", err)
	}
	fn := tmpFile.Name()
	os.Remove(fn)
	f, err := os.Create(fn)
	if err != nil {
		t.Fatalf(" cannot create %s: %v", fn, err)
	}
	f.Close()

	//watch non-existing file
	fw := file.Watcher(fn, file.Write)
	wrote := false
	go func() {
		<-fw
		wrote = true
	}()

	f, err = os.Create(fn)
	if err != nil {
		t.Fatalf(" cannot create %s: %v", fn, err)
	}
	f.Write([]byte{1, 2, 3})
	f.Close()

	for i := 0; i < 5; i++ {
		time.Sleep(time.Millisecond * 100)
		if wrote {
			break
		}
	}

	if !wrote {
		t.Fatalf("not written")
	}
}
