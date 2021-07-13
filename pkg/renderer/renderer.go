package renderer

import (
	"fmt"
	"io"
	"strings"

	"github.com/tsuzu/md2puki/pkg/urlutil"
	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/renderer"
)

var (
	unhandledError = fmt.Errorf("unhandled")
	skipElement    = fmt.Errorf("skip element")
)

func NewRenderer() renderer.Renderer {
	return &Renderer{}
}

type Renderer struct {
}

var _ renderer.Renderer = &Renderer{}

func (r *Renderer) renderChildren(src []byte, n ast.Node, hook func(n ast.Node, generated string) (string, error)) (string, error) {
	if !n.HasChildren() {
		return "", nil
	}

	rendered := make([]string, 0, n.ChildCount())

	err := ast.Walk(n, func(child ast.Node, entering bool) (ast.WalkStatus, error) {
		if n == child {
			return ast.WalkContinue, nil
		} else if !entering {
			return ast.WalkSkipChildren, nil
		}

		s, err := r.render(src, child)

		if hook != nil && err == nil {
			s, err = hook(child, s)
		}

		if err == skipElement {
			return ast.WalkSkipChildren, nil
		} else if err != nil {
			return ast.WalkSkipChildren, err
		}

		if child.Type() != ast.TypeInline {
			if child != n.FirstChild() {
				s = "\n" + s
			}

			if child.HasBlankPreviousLines() {
				s = "\n" + s
			}
		}

		rendered = append(rendered, s)

		return ast.WalkSkipChildren, nil
	})

	if err != nil {
		return "", err
	}

	return strings.Join(rendered, ""), nil
}

func (r *Renderer) autoLink(src []byte, n *ast.AutoLink) (string, error) {
	return "", unhandledError
}
func (r *Renderer) blockquote(src []byte, n *ast.Blockquote) (string, error) {
	s, err := r.renderChildren(src, n, nil)

	if err != nil {
		return "", err
	}

	return processLines(s, func(i int, s string) string {
		if i == 0 && len(s) == 0 {
			return s
		}

		if len(s) != 0 && s[0] == '>' {
			return ">" + s
		}

		return "> " + s
	}), nil
}
func (r *Renderer) codeBlock(src []byte, n *ast.CodeBlock) (string, error) {
	res := make([]string, 0, n.Lines().Len())
	for _, l := range n.Lines().Sliced(0, n.Lines().Len()) {
		res = append(res, " "+string(l.Value(src)))
	}

	return strings.Join(res, ""), nil
}
func (r *Renderer) codeSpan(src []byte, n *ast.CodeSpan) (string, error) {
	s, err := r.renderChildren(src, n, nil)

	if err != nil {
		return "", err
	}

	return `''` + s + `''`, nil
}
func (r *Renderer) document(src []byte, n *ast.Document) (string, error) {
	return "", unhandledError
}
func (r *Renderer) emphasis(src []byte, n *ast.Emphasis) (string, error) {
	s, err := r.renderChildren(src, n, nil)

	if err != nil {
		return "", err
	}

	if n.Level == 1 {
		return `'''` + s + `'''`, nil
	}

	return `''` + s + `''`, nil
}
func (r *Renderer) fencedCodeBlock(src []byte, n *ast.FencedCodeBlock) (string, error) {
	res := make([]string, 0, n.Lines().Len())
	for _, l := range n.Lines().Sliced(0, n.Lines().Len()) {
		res = append(res, " "+string(l.Value(src)))
	}

	return strings.Join(res, ""), nil
}
func (r *Renderer) htmlBlock(src []byte, n *ast.HTMLBlock) (string, error) {
	return "", unhandledError
}
func (r *Renderer) heading(src []byte, n *ast.Heading) (string, error) {
	level := n.Level
	if level > 3 {
		level = 3
	}

	g, err := r.renderChildren(src, n, nil)

	if err != nil {
		return "", err
	}

	return strings.Repeat("*", level) + " " + g, nil
}
func (r *Renderer) image(src []byte, n *ast.Image) (string, error) {
	return fmt.Sprintf("&ref(%s);", urlutil.EscapeURL(string(n.Destination))), nil
}
func (r *Renderer) link(src []byte, n *ast.Link) (string, error) {
	return fmt.Sprintf("[[%s:%s]]", string(n.Text(src)), urlutil.EscapeURL(string(n.Destination))), nil
}
func (r *Renderer) list(src []byte, n *ast.List) (string, error) {
	return "", unhandledError
}
func (r *Renderer) listItem(src []byte, n *ast.ListItem) (string, error) {
	list := n.Parent().(*ast.List)

	return r.renderChildren(src, n, func(n ast.Node, generated string) (string, error) {
		_, isList := n.(*ast.List)

		if !isList {
			if list.IsOrdered() {
				return "+ " + generated, nil
			}

			return "- " + generated, nil
		}

		return processLines(generated, func(i int, s string) string {
			if len(s) == 0 {
				return ""
			}

			return string(s[0]) + s
		}), nil
	})
}
func (r *Renderer) paragraph(src []byte, n *ast.Paragraph) (string, error) {
	return "", unhandledError
}
func (r *Renderer) rawHTML(src []byte, n *ast.RawHTML) (string, error) {
	return "", unhandledError
}
func (r *Renderer) stringNode(src []byte, n *ast.String) (string, error) {
	return "", unhandledError
}
func (r *Renderer) text(src []byte, n *ast.Text) (string, error) {
	s := string(n.Segment.Value(src))

	if n.HardLineBreak() || n.SoftLineBreak() {
		s = s + "\n"
	}

	return s, nil
}
func (r *Renderer) textBlock(src []byte, n *ast.TextBlock) (string, error) {
	return "", unhandledError
}
func (r *Renderer) thematicBreak(src []byte, n *ast.ThematicBreak) (string, error) {
	return "", skipElement
}

func (r *Renderer) definitionDescription(src []byte, n *east.DefinitionDescription) (string, error) {
	return "", unhandledError
}
func (r *Renderer) definitionList(src []byte, n *east.DefinitionList) (string, error) {
	return "", unhandledError
}
func (r *Renderer) definitionTerm(src []byte, n *east.DefinitionTerm) (string, error) {
	return "", unhandledError
}
func (r *Renderer) footnote(src []byte, n *east.Footnote) (string, error) {
	return "", unhandledError
}
func (r *Renderer) footnoteBacklink(src []byte, n *east.FootnoteBacklink) (string, error) {
	return "", unhandledError
}
func (r *Renderer) footnoteLink(src []byte, n *east.FootnoteLink) (string, error) {
	return "", unhandledError
}
func (r *Renderer) footnoteList(src []byte, n *east.FootnoteList) (string, error) {
	return "", unhandledError
}
func (r *Renderer) strikethrough(src []byte, n *east.Strikethrough) (string, error) {
	return "", unhandledError
}
func (r *Renderer) table(src []byte, n *east.Table) (string, error) {
	alignments := n.Alignments

	return r.renderChildren(src, n, func(n ast.Node, generated string) (string, error) {
		_, isHeader := n.(*east.TableHeader)

		idx := 0
		results := make([]string, 0, n.ChildCount())
		err := ast.Walk(n, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			cell, ok := n.(*east.TableCell)
			if !ok || !entering {
				return ast.WalkContinue, nil
			}

			defer func() {
				idx++
			}()

			align := alignments[idx]
			if a := cell.Alignment; a != east.AlignNone {
				align = a
			}

			generated, err := r.render(src, n)

			if err != nil {
				return ast.WalkStop, err
			}

			if align == east.AlignNone {
				results = append(results, "|"+generated)
			} else {
				results = append(results, fmt.Sprintf("|%s:%s", strings.ToUpper(align.String()), generated))
			}

			return ast.WalkSkipChildren, nil
		})

		if err != nil {
			return "", err
		}

		gen := strings.Join(results, "")

		if isHeader {
			return gen + "|h", nil
		}

		return gen + "|", nil
	})
}
func (r *Renderer) tableCell(src []byte, n *east.TableCell) (string, error) {
	return "", unhandledError
}
func (r *Renderer) tableHeader(src []byte, n *east.TableHeader) (string, error) {
	return "", unhandledError
}
func (r *Renderer) tableRow(src []byte, n *east.TableRow) (string, error) {
	return "", unhandledError
}
func (r *Renderer) taskCheckBox(src []byte, n *east.TaskCheckBox) (string, error) {
	return "", unhandledError
}

func (r *Renderer) render(src []byte, n ast.Node) (s string, err error) {
	switch n := n.(type) {
	case *ast.AutoLink:
		s, err = r.autoLink(src, n)
	case *ast.Blockquote:
		s, err = r.blockquote(src, n)
	case *ast.CodeBlock:
		s, err = r.codeBlock(src, n)
	case *ast.CodeSpan:
		s, err = r.codeSpan(src, n)
	case *ast.Document:
		s, err = r.document(src, n)
	case *ast.Emphasis:
		s, err = r.emphasis(src, n)
	case *ast.FencedCodeBlock:
		s, err = r.fencedCodeBlock(src, n)
	case *ast.HTMLBlock:
		s, err = r.htmlBlock(src, n)
	case *ast.Heading:
		s, err = r.heading(src, n)
	case *ast.Image:
		s, err = r.image(src, n)
	case *ast.Link:
		s, err = r.link(src, n)
	case *ast.List:
		s, err = r.list(src, n)
	case *ast.ListItem:
		s, err = r.listItem(src, n)
	case *ast.Paragraph:
		s, err = r.paragraph(src, n)
	case *ast.RawHTML:
		s, err = r.rawHTML(src, n)
	case *ast.String:
		s, err = r.stringNode(src, n)
	case *ast.Text:
		s, err = r.text(src, n)
	case *ast.TextBlock:
		s, err = r.textBlock(src, n)
	case *ast.ThematicBreak:
		s, err = r.thematicBreak(src, n)
	case *east.DefinitionDescription:
		s, err = r.definitionDescription(src, n)
	case *east.DefinitionList:
		s, err = r.definitionList(src, n)
	case *east.DefinitionTerm:
		s, err = r.definitionTerm(src, n)
	case *east.Footnote:
		s, err = r.footnote(src, n)
	case *east.FootnoteBacklink:
		s, err = r.footnoteBacklink(src, n)
	case *east.FootnoteLink:
		s, err = r.footnoteLink(src, n)
	case *east.FootnoteList:
		s, err = r.footnoteList(src, n)
	case *east.Strikethrough:
		s, err = r.strikethrough(src, n)
	case *east.Table:
		s, err = r.table(src, n)
	case *east.TableCell:
		s, err = r.tableCell(src, n)
	case *east.TableHeader:
		s, err = r.tableHeader(src, n)
	case *east.TableRow:
		s, err = r.tableRow(src, n)
	case *east.TaskCheckBox:
		s, err = r.taskCheckBox(src, n)
	}

	switch err {
	case nil:
		return s, nil
	case unhandledError:
		return r.renderChildren(src, n, nil)
	default:
		return "", err
	}
}

func (r *Renderer) Render(w io.Writer, source []byte, n ast.Node) error {
	s, err := r.render(source, n)

	if err != nil {
		return err
	}

	if len(s) != 0 && s[len(s)-1] != '\n' {
		s = s + "\n"
	}

	_, err = io.Copy(w, strings.NewReader(s))

	return err
}

// AddOptions adds given option to this renderer.
func (r *Renderer) AddOptions(...renderer.Option) {}

func processLines(s string, fn func(i int, s string) string) string {
	split := strings.Split(s, "\n")
	for i, s := range split {
		split[i] = fn(i, s)
	}

	return strings.Join(split, "\n")
}
