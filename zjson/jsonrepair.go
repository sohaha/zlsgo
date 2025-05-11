package zjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/sohaha/zlsgo/zstring"
)

type JSONSyntaxError struct {
	Message  string
	Position int
}

func (e *JSONSyntaxError) Error() string {
	return fmt.Sprintf("JSON syntax error at position %d: %s", e.Position, e.Message)
}

type RepairOptions struct {
	AllowComments       bool
	AllowTrailingCommas bool
	AllowSingleQuotes   bool
	AllowUnquotedKeys   bool
}

var defaultRepairOptions = &RepairOptions{
	AllowComments:       true,
	AllowTrailingCommas: true,
	AllowSingleQuotes:   true,
	AllowUnquotedKeys:   true,
}

func Repair(src string, opt ...func(*RepairOptions)) (dst string, err error) {
	opts := defaultRepairOptions
	for _, v := range opt {
		v(opts)
	}

	if src == "" {
		return `""`, nil
	}

	src = strings.TrimSpace(src)

	if src == "" {
		return `""`, nil
	}

	if strings.HasPrefix(src, "```json") {
		src = strings.TrimPrefix(src, "```json")
		if pos := strings.LastIndex(src, "```"); pos >= 0 {
			src = src[:pos]
		}
		src = strings.TrimSpace(src)
	} else if strings.HasSuffix(src, "```") {
		src = strings.TrimSuffix(src, "```")
		src = strings.TrimSpace(src)
	}

	if len(src) == 1 {
		c := src[0]
		switch c {
		case '{', '[':
			if c == '{' {
				return "{}", nil
			}
			return "[]", nil
		case '}', ']':
			return fmt.Sprintf(`"%c"`, c), nil
		case '"', '\'':
			return `""`, nil
		case ' ', '\t', '\n', '\r':
			return `""`, nil
		default:
			return fmt.Sprintf(`"%c"`, c), nil
		}
	}

	if !opts.AllowTrailingCommas {
		if strings.Contains(src, ",]") || strings.Contains(src, ",}") {
			pos1 := strings.Index(src, ",]")
			pos2 := strings.Index(src, ",}")

			var pos int
			if pos1 >= 0 && (pos2 < 0 || pos1 < pos2) {
				pos = pos1
			} else if pos2 >= 0 {
				pos = pos2
			} else {
				pos = -1
			}

			return "", &JSONSyntaxError{
				Position: pos,
				Message:  "trailing commas are not allowed",
			}
		}
	}

	srcBytes := zstring.String2Bytes(src)
	if json.Valid(srcBytes) {
		buf := &bytes.Buffer{}
		buf.Grow(len(src))
		if err = json.Compact(buf, srcBytes); err != nil {
			return "", err
		}
		dst = buf.String()
		return
	}

	bufSize := len(src)

	if opts.AllowComments {
		src = removeComments(src)
	}

	parser := newJSONParser(src, opts)
	result := parser.parseJSON()

	if parser.err != nil {
		return "", parser.err
	}

	buf := &bytes.Buffer{}
	buf.Grow(bufSize)
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal repaired JSON: %w", err)
	}

	dst = strings.TrimSuffix(buf.String(), "\n")
	return
}

func removeComments(src string) string {
	result := zstring.Buffer(len(src))

	inString := false
	stringChar := byte(0)
	escaped := false
	i := 0

	for i < len(src) {
		if inString {
			result.WriteByte(src[i])
			if src[i] == '\\' && !escaped {
				escaped = true
			} else if src[i] == stringChar && !escaped {
				inString = false
				stringChar = 0
			} else {
				escaped = false
			}
			i++
			continue
		}

		if src[i] == '"' || src[i] == '\'' {
			inString = true
			stringChar = src[i]
			result.WriteByte(src[i])
			i++
			continue
		}

		if i < len(src)-1 && src[i] == '/' && src[i+1] == '/' {
			i += 2
			for i < len(src) && src[i] != '\n' {
				i++
			}
			if i < len(src) {
				result.WriteByte('\n')
				i++
			}
			continue
		}

		if i < len(src)-1 && src[i] == '/' && src[i+1] == '*' {
			i += 2
			for i < len(src)-1 && !(src[i] == '*' && src[i+1] == '/') {
				if src[i] == '\n' {
					result.WriteByte('\n')
				}
				i++
			}
			if i < len(src)-1 {
				i += 2
			}
			continue
		}

		result.WriteByte(src[i])
		i++
	}

	return result.String()
}

type jsonParser struct {
	err       error
	options   *RepairOptions
	container string
	marker    []string
	index     int
}

func newJSONParser(in string, opts *RepairOptions) *jsonParser {
	if opts == nil {
		opts = defaultRepairOptions
	}
	return &jsonParser{
		container: in,
		index:     0,
		marker:    make([]string, 0, 8),
		options:   opts,
		err:       nil,
	}
}

func (p *jsonParser) parseJSON() interface{} {
	if len(p.container) == 0 {
		return ""
	}

	isPlainText := true
	for i := 0; i < len(p.container); i++ {
		c := p.container[i]
		if bytes.IndexByte([]byte{'{', '[', ':', ',', '"', '\'', '`'}, c) != -1 || zstring.IsSpace(rune(c)) {
			isPlainText = false
			break
		}
	}
	if isPlainText {
		return p.container
	}

	for {
		c, b := p.getByte(0)

		if !b {
			return ""
		}

		isInMarkers := len(p.marker) > 0

		switch {
		case c == '{':
			p.index++
			return p.parseObject()
		case c == '[':
			p.index++
			return p.parseArray()
		case c == '}':
			return ""
		case c == ']':
			return ""
		case isInMarkers && (c == '"' || c == '\'' || isLetter(c)):
			return p.parseString()
		case isInMarkers && (isNumber(c) || c == '-' || c == '.'):
			return p.parseNumber()
		}

		p.index++
	}
}

func (p *jsonParser) setError(message string) {
	if p.err == nil {
		p.err = &JSONSyntaxError{
			Position: p.index,
			Message:  message,
		}
	}
}

func (p *jsonParser) parseObject() map[string]interface{} {
	estimatedSize := 8
	if p.index < len(p.container) {
		end := strings.IndexByte(p.container[p.index:], '}')
		if end > 0 {
			segment := p.container[p.index : p.index+end]
			commaCount := strings.Count(segment, ",")
			estimatedSize = commaCount + 1
		}
	}

	rst := make(map[string]interface{}, estimatedSize)

	var c byte
	var b bool

	c, b = p.getByte(0)

	for b && c != '}' {
		p.skipWhitespaces()

		c, b = p.getByte(0)
		if b && c == ':' {
			p.index++
		}

		p.setMarker("object_key")
		p.skipWhitespaces()

		var key string
		_, b = p.getByte(0)
		for key == "" && b {
			currentIndex := p.index
			key = p.parseString().(string)

			c, b = p.getByte(0)
			if key == "" && b && c == ':' {
				key = "empty_placeholder"
				break
			} else if key == "" && p.index == currentIndex {
				p.index++
			}
		}

		p.skipWhitespaces()

		c, b = p.getByte(0)
		if b && c == '}' {
			prevC, prevB := p.getByte(-1)
			if prevB && prevC == ',' && !p.options.AllowTrailingCommas {
				p.setError("trailing commas are not allowed")
			}
			continue
		}

		p.skipWhitespaces()

		c, b = p.getByte(0)
		if !b || c != ':' {
			if b {
				p.setError(fmt.Sprintf("expected ':' after key '%s', got '%c'", key, c))
			}
		}

		p.index++
		p.resetMarker()
		p.setMarker("object_value")
		value := p.parseJSON()

		p.resetMarker()
		if key == "" && value == "" {
			continue
		}
		rst[key] = value

		c, b = p.getByte(0)
		if b && bytes.IndexByte([]byte{',', '\'', '"'}, c) != -1 {
			p.index++
		}

		p.skipWhitespaces()
		c, b = p.getByte(0)
	}

	if b && c == '}' {
		prevC, prevB := p.getByte(-1)
		if prevB && prevC == ',' && !p.options.AllowTrailingCommas {
			p.setError("trailing commas are not allowed")
		}
	} else if b {
		p.setError(fmt.Sprintf("expected '}' at end of object, got '%c'", c))
	}

	p.index++
	return rst
}

func (p *jsonParser) parseArray() []interface{} {
	estimatedSize := 8
	if p.index < len(p.container) {
		end := strings.IndexByte(p.container[p.index:], ']')
		if end > 0 {
			segment := p.container[p.index : p.index+end]
			commaCount := strings.Count(segment, ",")
			estimatedSize = commaCount + 1
		}
	}

	rst := make([]interface{}, 0, estimatedSize)

	var c byte
	var b bool

	p.setMarker("array")

	c, b = p.getByte(0)

	for b && c != ']' {
		p.skipWhitespaces()
		value := p.parseJSON()

		if value == nil || value == "" {
			break
		}

		if tc, ok := value.(string); ok && tc == "" {
			break
		}

		c, b = p.getByte(-1)
		if value == "..." && b && c == '.' {
		} else {
			rst = append(rst, value)
		}

		c, b = p.getByte(0)
		for b && (zstring.IsSpace(rune(c)) || c == ',') {
			p.index++
			c, b = p.getByte(0)
		}

		if p.getMarker() == "object_value" && c == '}' {
			break
		}
	}

	c, b = p.getByte(0)
	if b && c != ']' {
		if c == ',' && p.options.AllowTrailingCommas {
		} else {
			p.setError(fmt.Sprintf("expected ']' at end of array, got '%c'", c))
		}
		p.index--
	}

	p.index++
	p.resetMarker()
	return rst
}

func (p *jsonParser) parseString() interface{} {
	var missingQuotes, doubledQuotes bool
	var lStringDelimiter, rStringDelimiter byte = '"', '"'

	var c byte
	var b bool

	if len(p.container) == 1 {
		if p.container[0] == '"' || p.container[0] == '\'' {
			return ""
		}
	}

	c, b = p.getByte(0)
	for b && c != '"' && c != '\'' && !isLetter(c) {
		p.index++
		c, b = p.getByte(0)
	}

	if !b {
		return ""
	}

	switch {
	case c == '\'':
		if !p.options.AllowSingleQuotes {
			p.setError("single quotes are not allowed")
			return ""
		}
		lStringDelimiter = '\''
		rStringDelimiter = '\''
	case isLetter(c):
		if p.getMarker() == "object_key" && !p.options.AllowUnquotedKeys {
			p.setError("unquoted keys are not allowed")
			return ""
		}

		if (c == 't' || c == 'T' || c == 'f' || c == 'F' || c == 'n' || c == 'N') &&
			p.getMarker() != "object_key" {
			value := p.parseBooleanOrNull()
			if vs, ok := value.(string); !ok {
				return value
			} else if vs != "" {
				return vs
			}
		}

		missingQuotes = true
	}

	if !missingQuotes {
		p.index++
	}

	c, b = p.getByte(0)

	if b && c == lStringDelimiter {
		if p.index+1 < len(p.container) && p.container[p.index+1] == rStringDelimiter {
			doubledQuotes = true
			p.index++
		} else {
			i := 1
			nextC, nextB := p.getByte(i)
			for nextB && nextC == ' ' {
				i++
				nextC, nextB = p.getByte(i)
			}

			if nextB && (nextC == ',' || nextC == ']' || nextC == '}') {
			} else {
				p.index++
			}
		}
	}

	bufSize := 32
	if p.index < len(p.container) {
		end := strings.IndexByte(p.container[p.index:], rStringDelimiter)
		if end > 0 {
			bufSize = end + 1
		} else {
			bufSize = len(p.container) - p.index
		}
	}

	rst := make([]byte, 0, bufSize)

	c, b = p.getByte(0)

	marker := p.getMarker()

	for b && c != rStringDelimiter {
		if missingQuotes {
			if marker == "object_key" && (c == ':' || zstring.IsSpace(rune(c))) {
				break
			} else if marker == "object_value" && (c == ',' || c == '}') {
				break
			}
		}

		rst = append(rst, c)
		p.index++

		c, b = p.getByte(0)

		if len(rst) > 0 && rst[len(rst)-1] == '\\' {
			rst = rst[:len(rst)-1]
			if c == 't' {
				rst = append(rst, '\t')
				p.index++
			} else if c == 'n' {
				rst = append(rst, '\n')
				p.index++
			} else if c == 'r' {
				rst = append(rst, '\r')
				p.index++
			} else if c == 'b' {
				rst = append(rst, '\b')
				p.index++
			} else if c == '\\' || c == rStringDelimiter {
				rst = append(rst, c)
				p.index++
			}

			c, b = p.getByte(0)
		}

		if c == rStringDelimiter {
			if doubledQuotes && p.index+1 < len(p.container) && p.container[p.index+1] == rStringDelimiter {
				p.index++
				c, b = p.getByte(0)
				continue
			}

			if missingQuotes && marker == "object_value" {
				i := 1
				nextC, nextB := p.getByte(i)
				for nextB && zstring.IsSpace(rune(nextC)) {
					i++
					nextC, nextB = p.getByte(i)
				}

				if nextB && nextC == ':' {
					p.index--
					c, b = p.getByte(0)
					break
				}
			}
		}
	}

	if b && missingQuotes && marker == "object_key" && zstring.IsSpace(rune(c)) {
		p.skipWhitespaces()
		ci, bi := p.getByte(0)
		if !bi || (ci != ':' && ci != ',') {
			return ""
		}
	}

	if !b || c != rStringDelimiter {
		if missingQuotes && marker == "object_key" {
		} else if len(rst) > 0 {
			return zstring.Bytes2String(rst)
		} else {
			p.setError("unterminated string")
		}
	} else {
		p.index++
	}

	if len(rst) > 0 {
		i := len(rst) - 1
		for i >= 0 && zstring.IsSpace(rune(rst[i])) {
			i--
		}
		if i < len(rst)-1 {
			rst = rst[:i+1]
		}
	}

	return zstring.Bytes2String(rst)
}

func (p *jsonParser) parseNumber() interface{} {
	var rst []byte

	bufSize := 16
	if p.index < len(p.container) {
		i := p.index
		for i < len(p.container) && (isNumber(p.container[i]) ||
			p.container[i] == '-' || p.container[i] == '.' ||
			p.container[i] == 'e' || p.container[i] == 'E') {
			i++
		}
		bufSize = i - p.index + 1
	}

	rst = make([]byte, 0, bufSize)

	var c byte
	var b bool
	c, b = p.getByte(0)

	isArray := p.getMarker() == "array"

	for b && (isNumber(c) || c == '-' || c == '.' || c == 'e' || c == 'E' || c == '/' ||
		(c == ',' && !isArray)) {
		rst = append(rst, c)
		p.index++
		c, b = p.getByte(0)
	}

	if len(rst) > 1 {
		lastChar := rst[len(rst)-1]
		if lastChar == '-' || lastChar == 'e' || lastChar == 'E' || lastChar == '/' || lastChar == ',' {
			rst = rst[:len(rst)-1]
			p.index--
		}
	}

	switch {
	case len(rst) == 0:
		return p.parseJSON()
	case bytes.ContainsRune(rst, ','):
		return zstring.Bytes2String(rst)
	case bytes.ContainsRune(rst, '.') || bytes.ContainsRune(rst, 'e') || bytes.ContainsRune(rst, 'E'):
		r, err := strconv.ParseFloat(zstring.Bytes2String(rst), 64)
		if err != nil {
			p.setError(fmt.Sprintf("invalid number: %s", zstring.Bytes2String(rst)))
			return 0.0
		}
		return r
	case zstring.Bytes2String(rst) == "-":
		return p.parseJSON()
	}

	r, err := strconv.Atoi(zstring.Bytes2String(rst))
	if err != nil {
		p.setError(fmt.Sprintf("invalid number: %s", zstring.Bytes2String(rst)))
		return 0
	}
	return r
}

func (p *jsonParser) parseBooleanOrNull() interface{} {
	startingIndex := p.index

	var c byte
	var b bool
	c, b = p.getByte(0)
	c = byte(toLowerCase(c))

	if !b {
		return ""
	}

	switch c {
	case 't':
		if p.tryMatch("true") {
			return true
		}
	case 'f':
		if p.tryMatch("false") {
			return false
		}
	case 'n':
		if p.tryMatch("null") {
			return nil
		}
	}

	p.index = startingIndex
	return ""
}

func (p *jsonParser) tryMatch(target string) bool {
	if p.index+len(target) > len(p.container) {
		return false
	}

	for i := 0; i < len(target); i++ {
		c := p.container[p.index+i]
		if toLowerCase(c) != target[i] {
			return false
		}
	}

	p.index += len(target)
	return true
}

func (p *jsonParser) getByte(count int) (byte, bool) {
	if p.index+count < 0 || p.index+count >= len(p.container) {
		return ' ', false
	}

	return p.container[p.index+count], true
}

func (p *jsonParser) skipWhitespaces() {
	var c byte
	var b bool
	c, b = p.getByte(0)

	for b && zstring.IsSpace(rune(c)) {
		p.index++
		c, b = p.getByte(0)
	}
}

func (p *jsonParser) setMarker(in string) {
	if in != "" {
		p.marker = append(p.marker, in)
	}
}

func (p *jsonParser) resetMarker() {
	if len(p.marker) > 0 {
		p.marker = p.marker[:len(p.marker)-1]
	}
}

func (p *jsonParser) getMarker() string {
	if len(p.marker) > 0 {
		return p.marker[len(p.marker)-1]
	}

	return ""
}

func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isNumber(c byte) bool {
	return c >= '0' && c <= '9'
}

func toLowerCase(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + 32
	}
	return c
}
