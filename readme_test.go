package voiceml_test

// Compiles every ```go block in README.md against this checkout, so README
// examples cannot drift from the API. Complete programs (starting with
// "package main") build as-is; every other block is compiled as the body of
// a function that receives ctx and an existing client c.

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var goFence = regexp.MustCompile("(?s)```go\n(.*?)```")

const snippetPreamble = `package snippets

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	voiceml "github.com/voicetel/voiceml-go-sdk"
)

var (
	_ = errors.Is
	_ = fmt.Sprint
	_ = io.Copy
	_ = log.Fatal
	_ = http.StatusOK
	_ = os.Getenv
	_ = time.Second
)
`

func TestREADMESnippetsCompile(t *testing.T) {
	if testing.Short() {
		t.Skip("-short: skipping README compilation (it runs the go command)")
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go command not on PATH; cannot compile README snippets")
	}
	readme, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	repo, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	blocks := goFence.FindAllStringSubmatch(string(readme), -1)
	if len(blocks) < 10 {
		t.Fatalf("found only %d go blocks in README.md; the fence pattern is probably wrong", len(blocks))
	}

	dir := t.TempDir()
	write := func(name, content string) {
		t.Helper()
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", fmt.Sprintf("module readme.example\n\ngo 1.21\n\nrequire github.com/voicetel/voiceml-go-sdk v0.0.0\n\nreplace github.com/voicetel/voiceml-go-sdk => %s\n", repo))
	for i, m := range blocks {
		code := m[1]
		if strings.HasPrefix(strings.TrimSpace(code), "package main") {
			write(fmt.Sprintf("cmd%d/main.go", i), code)
			continue
		}
		var b strings.Builder
		if strings.Contains(code, ".All()") {
			b.WriteString("//go:build go1.23\n\n") // range-over-func
		}
		b.WriteString(snippetPreamble)
		fmt.Fprintf(&b, "\n// README block %d\nfunc snippet%d(ctx context.Context, c *voiceml.Client) {\n%s}\n", i, i, code)
		write(fmt.Sprintf("snippets/snippet%d.go", i), b.String())
	}

	cmd := exec.Command(goBin, "vet", "./...")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS=-mod=mod")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("README snippets do not compile:\n%s", out)
	}
}
