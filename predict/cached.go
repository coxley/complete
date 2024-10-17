package predict

import (
	"cmp"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/coxley/complete/args"
	"github.com/coxley/complete/cmplog"
)

const (
	// CachedDir is the parent directory under the user cache dir that cached suggestions
	// will be stored
	CachedDir = "tab_complete"
)

// Allow overriding for tests, but defaults to the operating system's preferred
// location.
var UserCacheDir = func() (string, error) {
	return os.UserCacheDir()
}

// Cached returns a predictor that can re-use previous values for some time
// until needing to regenerate.
//
// This is useful when the source of truth for items is expensive to calculate, like
// requiring a network request.
//
// Suggestions are written to a file in the preferred cache directory for the user's
// OS. The contents of the file a line-delimeted, with every entry being a suggestion.
// The file's modification time is used to determine whether to refresh or not.
//
// An example filepath might look like: ~/.cache/tab_complete/cmd/name
//
// Where 'cmd' is os.Args[0] and 'name' is provided to this function. [CachedDir] can
// be written to customize the containing directory.
func Cached(name string, ttl time.Duration, load func() []string) Predictor {
	return &cachePredictor{
		scope: os.Args[0],
		name:  name,
		ttl:   ttl,
		load:  load,
	}
}

// ScopedCache returns a cached predictor that is scoped manually instead of by the
// command.
//
// This is useful if you have multiple command-line programs that want to share
// tab-complete suggestions without duplicating the results.
func ScopedCache(scope string, name string, ttl time.Duration, load func() []string) Predictor {
	return &cachePredictor{
		scope: scope,
		name:  name,
		ttl:   ttl,
		load:  load,
	}
}

type cachedEntry struct {
	lastUpdate time.Time
	file       *os.File
	values     []string
}
type cachePredictor struct {
	load  func() []string
	scope string
	name  string
	ttl   time.Duration
}

func (p *cachePredictor) loadCache() (*cachedEntry, error) {
	f, err := p.open()
	if err != nil {
		return nil, err
	}

	// Format is stored as line-delimited string values
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	// Force cache filling on no entries
	if len(b) == 0 {
		return &cachedEntry{file: f}, nil
	}

	values := make([]string, 0, 128)
	for _, v := range strings.Split(string(b), "\n") {
		if v == "" {
			continue
		}
		values = append(values, v)
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	return &cachedEntry{
		file:   f,
		values: values,
		// Use mtime as last update
		lastUpdate: fi.ModTime(),
	}, nil
}

func (p *cachePredictor) refresh(f *os.File) ([]string, error) {
	// Should the fill function return nothing, leave stored values unchanged to
	// prevent intermittent issue wiping what we have.
	values := p.load()
	if len(values) == 0 {
		cmplog.Log("cached pred %s:%s returned no results", p.scope, p.name)
		return nil, nil
	}

	slices.Sort(values)

	// The passed in file can't be opened with O_TRUNC to avoid setting the modtime
	// before we have a chance to read it ourselves.
	if err := f.Truncate(0); err != nil {
		return nil, fmt.Errorf("truncating %q: %w", f.Name(), err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("seeking %q: %w", f.Name(), err)
	}
	if _, err := fmt.Fprintln(f, strings.Join(values, "\n")); err != nil {
		return nil, fmt.Errorf("writing %q: %w", f.Name(), err)
	}
	return values, nil
}

func (p *cachePredictor) Predict(args args.Args) []string {
	entry, err := p.loadCache()
	if err != nil {
		cmplog.Log("cached pred %s:%s failed to load cache: %v", p.scope, p.name, err)
		return nil
	}

	values := entry.values

	// Refresh the cache
	if time.Since(entry.lastUpdate) > p.ttl {
		values, err = p.refresh(entry.file)
		if err != nil {
			cmplog.Log("cached pred %s:%s failed to refresh: %v", p.scope, p.name, err)
			return values
		}
	}

	// Cursor is in an empty space
	if args.Last == "" {
		return values
	}

	var suggestions []string
	for _, text := range values {
		if strings.HasPrefix(text, args.Last) {
			suggestions = append(suggestions, text)
		}
	}
	return suggestions
}

// filepath returns the path to the suggestions file
func (p *cachePredictor) filepath() (string, error) {
	cache, err := UserCacheDir()
	if err != nil {
		return "", err
	}

	// Directory defaults to being specific for the current command-line tool
	// to prevent conflicting names for different setups.
	//
	// Providing a custom scope puts it in the hands of the tool writer.
	parentDir := filepath.Join(cache, CachedDir, cmp.Or(p.scope, os.Args[0]))
	return filepath.Join(parentDir, p.name), nil
}

// open returns a file handle to the suggestions file, creating if necessary.
func (p *cachePredictor) open() (*os.File, error) {
	path, err := p.filepath()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	return os.OpenFile(
		path,
		os.O_RDWR|os.O_CREATE,
		0o644,
	)
}
