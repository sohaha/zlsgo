package zstring

import (
	"bytes"
	"fmt"
	"io"
)

// Template implements a simple template engine that replaces tags in a template string.
// It supports custom start and end tag delimiters and efficient processing.
type Template struct {
	template string // The original template string
	startTag []byte // Byte representation of the opening tag delimiter
	endTag   []byte // Byte representation of the closing tag delimiter

	texts [][]byte // Slices of text between tags
	tags  []string // The tag names extracted from the template
}

// NewTemplate creates a new template with the specified template string and tag delimiters.
// It parses the template and returns an error if the template format is invalid.
func NewTemplate(template, startTag, endTag string) (*Template, error) {
	t := Template{
		startTag: String2Bytes(startTag),
		endTag:   String2Bytes(endTag),
	}

	err := t.ResetTemplate(template)

	return &t, err
}

// ResetTemplate changes the template string and re-parses it.
// Returns an error if the template format is invalid (e.g., missing end tags).
func (t *Template) ResetTemplate(template string) error {
	t.template = template
	t.tags = t.tags[:0]
	t.texts = t.texts[:0]

	s := String2Bytes(template)

	tagsCount := bytes.Count(s, t.startTag)
	if tagsCount == 0 {
		return nil
	}

	if tagsCount+1 > cap(t.texts) {
		t.texts = make([][]byte, 0, tagsCount+1)
	}
	if tagsCount > cap(t.tags) {
		t.tags = make([]string, 0, tagsCount)
	}

	for {
		n := bytes.Index(s, t.startTag)
		if n < 0 {
			t.texts = append(t.texts, s)
			break
		}
		t.texts = append(t.texts, s[:n])

		s = s[n+len(t.startTag):]
		n = bytes.Index(s, t.endTag)
		if n < 0 {
			return fmt.Errorf("cannot find end tag=%q in the template=%q starting from %q", Bytes2String(t.endTag), template, s)
		}

		t.tags = append(t.tags, Bytes2String(s[:n]))
		s = s[n+len(t.endTag):]
	}

	return nil
}

// Process executes the template, writing the result to the provided writer.
// For each tag encountered, it calls the provided function with the tag name.
// Returns the number of bytes written and any error encountered.
func (t *Template) Process(w io.Writer, fn func(w io.Writer, tag string) (int, error)) (int64, error) {
	var nn int64
	n := len(t.texts) - 1
	if n == -1 {
		ni, err := w.Write(String2Bytes(t.template))
		return int64(ni), err
	}

	for i := 0; i < n; i++ {
		ni, err := w.Write(t.texts[i])
		nn += int64(ni)
		if err != nil {
			return nn, err
		}

		ni, err = fn(w, t.tags[i])
		nn += int64(ni)
		if err != nil {
			return nn, err
		}
	}
	ni, _ := w.Write(t.texts[n])
	nn += int64(ni)
	return nn, nil
}
