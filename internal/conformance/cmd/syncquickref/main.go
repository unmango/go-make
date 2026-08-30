// Command syncquickref refreshes internal/conformance/testdata/quickref.json
// from the GNU Make manual. Run it with `make sync-quickref`.
//
// Two local HTML files may be passed instead, in the order Quick Reference,
// Special Built-in Target Names, to regenerate the fixture without network
// access.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/unmango/go-make/internal/conformance"
)

const output = "internal/conformance/testdata/quickref.json"

var client = &http.Client{Timeout: 30 * time.Second}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	read := fetch
	args := os.Args[1:]
	if len(args) == 2 {
		read = os.ReadFile
		args = []string{args[0], args[1]}
	} else if len(args) > 0 {
		return fmt.Errorf("usage: syncquickref [quick-reference.html special-targets.html]")
	} else {
		args = []string{conformance.QuickRefURL, conformance.SpecialTargetsURL}
	}

	quickRef, err := read(args[0])
	if err != nil {
		return err
	}
	specialTargets, err := read(args[1])
	if err != nil {
		return err
	}

	ref, err := conformance.Parse(quickRef, specialTargets)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(ref, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.FromSlash(output), append(data, '\n'), 0o644); err != nil {
		return err
	}

	fmt.Printf("%s: %d directives, %d functions, %d special variables, %d special targets\n",
		output, len(ref.Directives), len(ref.Functions), len(ref.SpecialVariables), len(ref.SpecialTargets))
	return nil
}

func fetch(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	// gnu.org rejects the default Go user agent.
	req.Header.Set("User-Agent", "go-make-syncquickref (+https://github.com/unmango/go-make)")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", url, res.Status)
	}

	return io.ReadAll(res.Body)
}
