package zstring

const defFilterMask = '*'

type (
	scope struct {
		start int
		stop  int
	}
	node struct {
		children map[rune]*node
		end      bool
	}
	filterNode struct {
		node
		mask rune
	}
	replacer struct {
		node
		mapping map[string]string
	}
)

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

func (n *filterNode) Find(str string) []string {
	chars := []rune(str)
	if len(chars) == 0 {
		return nil
	}

	scopes := n.findKeywordScopes(chars)
	return n.collectKeywords(chars, scopes)
}

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

func (n *filterNode) replaceWithAsterisk(chars []rune, start, stop int) {
	for i := start; i < stop; i++ {
		chars[i] = n.mask
	}
}

func (n *filterNode) collectKeywords(chars []rune, scopes []scope) []string {
	set := make(map[string]struct{})
	for _, v := range scopes {
		set[string(chars[v.start:v.stop])] = struct{}{}
	}

	var i int
	keywords := make([]string, len(set))
	for k := range set {
		keywords[i] = k
		i++
	}

	return keywords
}

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

func NewReplacer(mapping map[string]string) *replacer {
	var rep = &replacer{
		mapping: mapping,
	}
	for k := range mapping {
		rep.add(k)
	}

	return rep
}

func (r *replacer) Replace(text string) string {
	var builder = Buffer()
	var chars = []rune(text)
	var size = len(chars)
	var start = -1

	for i := 0; i < size; i++ {
		child, ok := r.children[chars[i]]
		if !ok {
			builder.WriteRune(chars[i])
			continue
		}

		if start < 0 {
			start = i
		}
		var end = -1
		if child.end {
			end = i + 1
		}

		var j = i + 1
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
