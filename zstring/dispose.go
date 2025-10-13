package zstring

// Default character used to mask filtered words
const defFilterMask = '*'

type (
	// scope represents a substring range with start and stop indices
	scope struct {
		start int
		stop  int
	}
	// node is a trie node used for efficient string matching
	node struct {
		children map[rune]*node
		end      bool // indicates if this node is the end of a word
	}
	// filterNode extends node with text filtering capabilities
	filterNode struct {
		node
		mask rune // character used to replace filtered content
	}
	// replacer implements string replacement using a trie structure
	replacer struct {
		mapping map[string]string // maps original strings to their replacements
		node
	}
)

// NewFilter creates a new text filter that can identify and mask sensitive words.
// It accepts a list of words to filter and an optional mask character (defaults to '*').
func NewFilter(words []string, mask ...rune) *filterNode {
	n := &filterNode{
		mask: defFilterMask,
	}
	if len(mask) > 0 {
		n.mask = mask[0]
	}
	for _, word := range words {
		n.add(word)
	}

	return n
}

// add inserts a word into the trie structure.
func (n *node) add(word string) {
	chars := []rune(word)
	if len(chars) == 0 {
		return
	}

	nd := n
	for _, char := range chars {
		if nd.children == nil {
			child := new(node)
			nd.children = map[rune]*node{
				char: child,
			}
			nd = child
		} else if child, ok := nd.children[char]; ok {
			nd = child
		} else {
			child := new(node)
			nd.children[char] = child
			nd = child
		}
	}

	nd.end = true
}

// Find searches for all filtered words within the given string and returns them.
func (n *filterNode) Find(str string) []string {
	chars := []rune(str)
	if len(chars) == 0 {
		return nil
	}

	scopes := n.findKeywordScopes(chars)
	return n.collectKeywords(chars, scopes)
}

// Filter replaces all occurrences of filtered words with the mask character.
// It returns the filtered string, a list of found keywords, and a boolean indicating if any keywords were found.
func (n *filterNode) Filter(str string) (res string, keywords []string, found bool) {
	chars := []rune(str)
	if len(chars) == 0 {
		return str, nil, false
	}

	scopes := n.findKeywordScopes(chars)
	keywords = n.collectKeywords(chars, scopes)

	for _, match := range scopes {
		// we don't care about overlaps, not bringing a performance improvement
		n.replaceWithAsterisk(chars, match.start, match.stop)
	}

	return string(chars), keywords, len(keywords) > 0
}

// replaceWithAsterisk replaces characters in the given range with the mask character.
func (n *filterNode) replaceWithAsterisk(chars []rune, start, stop int) {
	for i := start; i < stop; i++ {
		chars[i] = n.mask
	}
}

// collectKeywords extracts unique keywords from the identified scopes in the text.
func (n *filterNode) collectKeywords(chars []rune, scopes []scope) []string {
	set := make(map[string]struct{}, len(scopes))
	for _, v := range scopes {
		set[string(chars[v.start:v.stop])] = struct{}{}
	}

	keywords := make([]string, 0, len(set))
	for k := range set {
		keywords = append(keywords, k)
	}

	return keywords
}

// findKeywordScopes identifies all occurrences of filtered words in the text
// and returns their position ranges.
func (n *filterNode) findKeywordScopes(chars []rune) []scope {
	var scopes []scope
	size := len(chars)
	start := -1

	for i := 0; i < size; i++ {
		child, ok := n.children[chars[i]]
		if !ok {
			continue
		}

		if start < 0 {
			start = i
		}
		if child.end {
			scopes = append(scopes, scope{
				start: start,
				stop:  i + 1,
			})
		}

		for j := i + 1; j < size; j++ {
			cchild, ok := child.children[chars[j]]
			if !ok {
				break
			}

			child = cchild
			if child.end {
				scopes = append(scopes, scope{
					start: start,
					stop:  j + 1,
				})
			}
		}

		start = -1
	}

	return scopes
}

// NewReplacer creates a new string replacer that efficiently replaces multiple strings at once.
// It uses a trie structure for fast matching.
func NewReplacer(mapping map[string]string) *replacer {
	rep := &replacer{
		mapping: mapping,
	}
	for k := range mapping {
		rep.add(k)
	}

	return rep
}

// Replace performs all configured string replacements on the input text.
// It efficiently identifies and replaces all occurrences of the mapped strings.
func (r *replacer) Replace(text string) string {
	builder := Buffer()
	chars := []rune(text)
	size := len(chars)
	start := -1

	for i := 0; i < size; i++ {
		child, ok := r.children[chars[i]]
		if !ok {
			builder.WriteRune(chars[i])
			continue
		}

		if start < 0 {
			start = i
		}
		end := -1
		if child.end {
			end = i + 1
		}

		j := i + 1
		for ; j < size; j++ {
			cchild, ok := child.children[chars[j]]
			if !ok {
				break
			}

			child = cchild
			if child.end {
				end = j + 1
				i = j
			}
		}

		if end > 0 {
			i = j - 1
			builder.WriteString(r.mapping[string(chars[start:end])])
		} else {
			builder.WriteRune(chars[i])
		}
		start = -1
	}

	return builder.String()
}
