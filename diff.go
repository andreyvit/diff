package diff

import (
	"bytes"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

func diff(a, b string) []diffmatchpatch.Diff {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(a, b, true)
	if len(diffs) > 2 {
		diffs = dmp.DiffCleanupSemantic(diffs)
		diffs = dmp.DiffCleanupEfficiency(diffs)
	}
	return diffs
}

// CharacterDiff returns an inline diff between the two strings, using (++added++) and (~~deleted~~) markup.
func CharacterDiff(a, b string) string {
	return diffsToString(diff(a, b))
}

func diffsToString(diffs []diffmatchpatch.Diff) string {
	var buff bytes.Buffer
	for _, diff := range diffs {
		text := diff.Text
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			buff.WriteString("(++")
			buff.WriteString(text)
			buff.WriteString("++)")
		case diffmatchpatch.DiffDelete:
			buff.WriteString("(~~")
			buff.WriteString(text)
			buff.WriteString("~~)")
		case diffmatchpatch.DiffEqual:
			buff.WriteString(text)
		}
	}
	return buff.String()
}

// LineDiff returns a normal linewise diff between the two given strings.
func LineDiff(a, b string) string {
	return strings.Join(LineDiffAsLines(a, b), "\n")
}

// LineDiffAsLines returns the lines of a linewise diff between the two given strings.
func LineDiffAsLines(a, b string) []string {
	return diffsToPatchLines(diff(a, b))
}

// LineDiffNoWhitespace returns a linewise diff between the two given strings
// with whitespace ignored. Leading, trailing, and repeated whitespace on each
// line does not count as a change, and an empty string is returned when the
// inputs differ only in whitespace. When there are other differences, the diff
// is rendered from the whitespace-normalized lines.
func LineDiffNoWhitespace(a, b string) string {
	return strings.Join(LineDiffAsLinesNoWhitespace(a, b), "\n")
}

// LineDiffAsLinesNoWhitespace returns the lines of a whitespace-insensitive
// linewise diff between the two given strings.
func LineDiffAsLinesNoWhitespace(a, b string) []string {
	a, b = normalizeWhitespace(a), normalizeWhitespace(b)
	if a == b {
		return nil
	}
	return diffsToPatchLines(diff(a, b))
}

// normalizeWhitespace trims each line and collapses runs of whitespace within
// a line to a single space.
func normalizeWhitespace(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for i, line := range lines {
		lines[i] = strings.Join(strings.Fields(line), " ")
	}
	return strings.Join(lines, "\n")
}

type patchBuilder struct {
	output        []string
	oldLines      []string
	newLines      []string
	newLineBuffer bytes.Buffer
	oldLineBuffer bytes.Buffer
}

func (b *patchBuilder) AddCharacters(text string, op diffmatchpatch.Operation) {
	if op == diffmatchpatch.DiffInsert || op == diffmatchpatch.DiffEqual {
		b.newLineBuffer.WriteString(text)
	}
	if op == diffmatchpatch.DiffDelete || op == diffmatchpatch.DiffEqual {
		b.oldLineBuffer.WriteString(text)
	}
}
func (b *patchBuilder) AddNewline(op diffmatchpatch.Operation) {
	oldLine := b.oldLineBuffer.String()
	newLine := b.newLineBuffer.String()

	if op == diffmatchpatch.DiffEqual && (oldLine == newLine) {
		b.FlushChunk()
		b.output = append(b.output, " "+newLine)
		b.oldLineBuffer.Reset()
		b.newLineBuffer.Reset()
	} else {
		if op == diffmatchpatch.DiffDelete || op == diffmatchpatch.DiffEqual {
			b.oldLines = append(b.oldLines, "-"+oldLine)
			b.oldLineBuffer.Reset()
		}
		if op == diffmatchpatch.DiffInsert || op == diffmatchpatch.DiffEqual {
			b.newLines = append(b.newLines, "+"+newLine)
			b.newLineBuffer.Reset()
		}
	}
}
func (b *patchBuilder) FlushChunk() {
	if b.oldLines != nil {
		b.output = append(b.output, b.oldLines...)
		b.oldLines = nil
	}
	if b.newLines != nil {
		b.output = append(b.output, b.newLines...)
		b.newLines = nil
	}
}
func (b *patchBuilder) Flush() {
	if b.oldLineBuffer.Len() > 0 && b.newLineBuffer.Len() > 0 {
		b.AddNewline(diffmatchpatch.DiffEqual)
	} else if b.oldLineBuffer.Len() > 0 {
		b.AddNewline(diffmatchpatch.DiffDelete)
	} else if b.newLineBuffer.Len() > 0 {
		b.AddNewline(diffmatchpatch.DiffInsert)
	}
	b.FlushChunk()
}

func diffsToPatchLines(diffs []diffmatchpatch.Diff) []string {
	b := new(patchBuilder)
	b.output = make([]string, 0, len(diffs))

	for _, diff := range diffs {
		lines := strings.Split(diff.Text, "\n")
		for idx, line := range lines {
			if idx > 0 {
				b.AddNewline(diff.Type)
			}
			b.AddCharacters(line, diff.Type)
		}
	}

	b.Flush()
	return b.output
}
