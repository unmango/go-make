// Command syncquickref writes internal/conformance/testdata/quickref.json from
// local copies of the GNU Make manual.
//
// The manual pages are content-addressed by nix, see the sync-quickref app in
// flake.nix, so this command performs no network access of its own. Run it with
// `make sync-quickref`.
//
// Usage: syncquickref <quick-reference.html> <special-targets.html>
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/unmango/go-make/internal/conformance"
)

const output = "internal/conformance/testdata/quickref.json"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: syncquickref <quick-reference.html> <special-targets.html>")
	}

	quickRef, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	specialTargets, err := os.ReadFile(args[1])
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
