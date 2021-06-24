package renderer

import (
	"bytes"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

func open(t *testing.T, path string) []byte {
	t.Helper()

	b, err := os.ReadFile(path)

	if err != nil {
		t.Fatal(err)
	}

	return b
}

func parse(t *testing.T, b []byte) ast.Node {
	t.Helper()

	md := goldmark.New(
		goldmark.WithExtensions(extension.NewTable()),
	)

	return md.Parser().Parse(text.NewReader(b))
}

func openString(t *testing.T, path string) string {
	t.Helper()

	return string(open(t, path))
}

func TestRenderer_Render(t *testing.T) {
	type args struct {
		source []byte
		n      ast.Node
	}
	initArgs := func(p string) args {
		b := open(t, p)

		return args{
			source: b,
			n:      parse(t, b),
		}
	}

	tests := []struct {
		name    string
		args    args
		wantW   string
		wantErr bool
	}{
		{
			name:    "Computer Architecture A",
			args:    initArgs("./testfiles/normal/input.md"),
			wantW:   openString(t, "./testfiles/normal/wanted.pukiwiki"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Renderer{}
			w := &bytes.Buffer{}
			if err := r.Render(w, tt.args.source, tt.args.n); (err != nil) != tt.wantErr {
				t.Errorf("Renderer.Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotW := w.String()
			if diff := cmp.Diff(tt.wantW, gotW); diff != "" {
				t.Errorf("Renderer.Render() = %s", diff)
			}
		})
	}
}
