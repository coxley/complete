package internal

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/coxley/complete/cmplog"
)

var (
	onceChdir sync.Once
	onceLog   sync.Once
)

func SetupLogging() {
	onceLog.Do(func() {
		// Set debug environment variable so logs will be printed
		if testing.Verbose() {
			os.Setenv(cmplog.Env, "1")
			// refresh the logger with environment variable set
			cmplog.Reset()
		}
	})
}

// Chdir into a well-known temporary structure, mostly useful for internal tests.
//
// Logic only executes once and stays changed.
//
// > tree -a
// ├── .dot.txt
// ├── a.txt
// ├── b.txt
// ├── c.txt
// ├── dir
// │ ├── bar
// │ └── foo
// ├── outer
// │ └── inner
// │     └── readme.md
// └── readme.md
func Chdir(t testing.TB) {
	SetupLogging()
	onceChdir.Do(func() {
		root, err := os.MkdirTemp("", "cmptest-*")
		if err != nil {
			t.Fatalf("creating dir: %v", err)
		}

		_, err = os.Create(filepath.Join(root, ".dot.txt"))
		if err != nil {
			t.Fatalf("creating file: %v", err)
		}

		_, err = os.Create(filepath.Join(root, "a.txt"))
		if err != nil {
			t.Fatalf("creating file: %v", err)
		}

		_, err = os.Create(filepath.Join(root, "b.txt"))
		if err != nil {
			t.Fatalf("creating file: %v", err)
		}

		_, err = os.Create(filepath.Join(root, "c.txt"))
		if err != nil {
			t.Fatalf("creating file: %v", err)
		}

		_, err = os.Create(filepath.Join(root, "readme.md"))
		if err != nil {
			t.Fatalf("creating file: %v", err)
		}

		dir := filepath.Join(root, "dir")
		err = os.Mkdir(dir, 0o700)
		if err != nil {
			t.Fatalf("creating dir: %v", err)
		}

		_, err = os.Create(filepath.Join(dir, "bar"))
		if err != nil {
			t.Fatalf("creating file: %v", err)
		}

		_, err = os.Create(filepath.Join(dir, "foo"))
		if err != nil {
			t.Fatalf("creating file: %v", err)
		}

		outer := filepath.Join(root, "outer")
		err = os.Mkdir(outer, 0o700)
		if err != nil {
			t.Fatalf("creating dir: %v", err)
		}

		inner := filepath.Join(outer, "inner")
		err = os.Mkdir(inner, 0o700)
		if err != nil {
			t.Fatalf("creating dir: %v", err)
		}

		_, err = os.Create(filepath.Join(inner, "readme.md"))
		if err != nil {
			t.Fatalf("creating file: %v", err)
		}

		err = os.Chdir(root)
		if err != nil {
			t.Fatalf("changing dir: %v", err)
		}
	})
}
