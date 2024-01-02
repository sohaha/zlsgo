package zstring

import (
	"bytes"
	"fmt"
	"io"
)

type Template struct {
	template string
	startTag []byte
	endTag   []byte

	texts [][]byte
	tags  []string
}

func NewTemplate(template, startTag, endTag string) (*Template, error) {
	t := Template{
		startTag: String2Bytes(startTag),
		endTag:   String2Bytes(endTag),
	}

	err := t.ResetTemplate(template)

	return &t, err
}

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
