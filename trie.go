package strings

import (
	"sort"
	"unsafe"
)

type PatriciaTrie struct {
	root *node
	Size int
}

func NewPatriciaTrie() *PatriciaTrie {
	return &PatriciaTrie{root: &node{}}
}

func (t *PatriciaTrie) estimateSize() uint64 {
	i := uint64(unsafe.Sizeof(t))
	i += t.root.estimateSize()
	return i
}

func (t *PatriciaTrie) Get(key string) interface{} {
	if len(key) == 0 {
		return nil
	}

	keyBytes := []byte(key)
	v := &visitor{}
	v.traverse(keyBytes, t.root)
	return v.result
}

func (t *PatriciaTrie) Contains(key string) bool {
	return t.Get(key) != nil
}

// todo an iterable or channel or stream based interface
func (t *PatriciaTrie) SearchPrefix(prefix string) []interface{} {
	if len(prefix) == 0 {
		return nil
	}

	keyBytes := []byte(prefix)
	n := search(keyBytes, t.root)
	if n == nil {
		return nil
	}

	matches := []*node{n}
	matches = append(matches, getNodes(n)...)

	var values []interface{}
	for _, match := range matches {
		if match.value != nil {
			values = append(values, match.value)
		}
	}
	return values
}

func (t *PatriciaTrie) ContainsPrefix(prefix string) bool {
	return len(t.SearchPrefix(prefix)) > 0
}

func (t *PatriciaTrie) Put(key string, value interface{}) {
	inserted, _ := t.insert([]byte(key), t.root, value)
	if inserted {
		t.Size++
	}
}

func (t *PatriciaTrie) Entries() map[string]interface{} {
	m := make(map[string]interface{})
	for _, n := range getNodes(t.root) {
		if n.isReal() {
			var k []byte
			for nn := n; nn != nil; nn = nn.parent {
				k = append(nn.key[:], k...)
			}
			m[string(k)] = n.value
		}
	}
	return m
}

func (t *PatriciaTrie) insert(key []byte, n *node, value interface{}) (inserted, updated bool) {
	prefixLength := n.getCommonPrefixLength(key)
	if t.root == n || prefixLength == 0 || (prefixLength < len(key) && prefixLength >= len(n.key)) {
		newKey := key[prefixLength:]
		child := n.findChild(newKey[0])
		if child != nil {
			return t.insert(newKey, child, value)
		}

		newNode := &node{key: newKey, value: value}
		newBranchNode := n.addChild(newNode)
		if newBranchNode != n && n.parent != nil {
			n.parent.removeChild(n).addChild(newBranchNode)
		}
		return true, false
	}

	if prefixLength == len(key) && prefixLength == len(n.key) {
		if n.value == nil {
			inserted = true
		}
		n.value = value
		return inserted, !inserted
	}

	if prefixLength > 0 && prefixLength < len(n.key) {
		suffixBytes := n.getKeySuffix(prefixLength)
		prefixBytes := append([]byte(nil), key[:prefixLength]...)

		rewrittenNode := &node{key: prefixBytes}

		if prefixLength == len(key) {
			rewrittenNode.value = value
		} else {
			newNode := &node{key: key[prefixLength:], value: value}
			rewrittenNode = rewrittenNode.addChild(newNode)
		}

		parentNode := n.parent
		if parentNode != nil {
			parentNode = parentNode.removeChild(n).addChild(rewrittenNode)
		}

		n.key = suffixBytes
		rewrittenNode.addChild(n)
		return true, false
	}

	child := &node{
		key:      n.getKeySuffix(prefixLength),
		children: n.children[:],
		value:    n.value,
	}

	n.key = key
	n.value = value
	n = n.addChild(child)

	return true, false
}

func search(keyBytes []byte, n *node) *node {
	prefixLength := n.getCommonPrefixLength(keyBytes)
	if prefixLength == len(keyBytes) && prefixLength <= len(n.key) {
		return n
	}

	if len(n.key) == 0 || (prefixLength < len(keyBytes) && prefixLength >= len(n.key)) {
		newKey := keyBytes[prefixLength:]
		child := n.findChild(newKey[0])
		if child != nil {
			return search(newKey, child)
		}
	}

	return nil
}

func getNodes(parent *node) []*node {
	var nodes []*node
	if parent != nil {
		nodes = append(nodes, parent.children...)
	}

	var result []*node
	for len(nodes) > 0 {
		next := nodes[0]
		nodes = nodes[1:]
		nodes = append(nodes, next.children...)
		if next.isReal() {
			result = append(result, next)
		}
	}
	return result
}

type node struct {
	parent   *node
	children []*node

	key   []byte
	value interface{}
}

func (n *node) estimateSize() uint64 {
	i := uint64(unsafe.Sizeof(n))
	i += uint64(len(n.key))
	for _, c := range n.children {
		i += c.estimateSize()
	}
	return i
}

func (n *node) isReal() bool {
	return n != nil && n.value != nil
}

func (n *node) getCommonPrefixLength(keyBytes []byte) int {
	key := n.key
	prefixLength := 0
	for prefixLength < len(keyBytes) && prefixLength < len(key) {
		if keyBytes[prefixLength] != key[prefixLength] {
			return prefixLength
		}
		prefixLength++
	}
	return prefixLength
}

func (n *node) compareTo(o *node) int {
	switch {
	case n == o:
		return 0
	case n == nil:
		return -1
	case o == nil:
		return 1
	}

	thisKeyBytes := n.key
	otherKeyBytes := o.key

	for i := 0; i < len(thisKeyBytes) && i < len(otherKeyBytes); i++ {
		thisByte, otherByte := thisKeyBytes[i], otherKeyBytes[i]
		switch {
		case thisByte < otherByte:
			return -1
		case thisByte > otherByte:
			return 1
		}
	}

	return len(thisKeyBytes) - len(otherKeyBytes)
}

func (n *node) addChild(child *node) *node {
	n.children = append(n.children, child)
	sort.Sort(nodeSort(n.children))
	child.parent = n
	return n
}

func (n *node) removeChild(child *node) *node {
	i := sort.Search(len(n.children), func(i int) bool { return string(n.children[i].key) >= string(child.key) })
	if i < len(n.children) && string(n.children[i].key) == string(child.key) {
		filtered := n.children[:i]
		if i+1 < len(n.children) {
			filtered = append(filtered, n.children[i+1:]...)
		}
		n.children = filtered
	}
	return n
}

type nodeSort []*node

func (a nodeSort) Len() int           { return len(a) }
func (a nodeSort) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a nodeSort) Less(i, j int) bool { return a[i].compareTo(a[j]) < 0 }

func (n node) String() string {
	return string(n.key)
}

func (n node) keyStartsWith(b byte) bool {
	return len(n.key) > 0 && n.key[0] == b
}

func (n node) getKeySuffix(offset int) []byte {
	return n.key[offset:]
}

func (n node) findChild(b byte) *node {
	i := sort.Search(len(n.children), func(i int) bool { return n.children[i].key[0] >= b })
	if i < len(n.children) && n.children[i].keyStartsWith(b) {
		return n.children[i]
	}

	return nil
}

type visitor struct {
	result interface{}
}

func (v *visitor) visit(n *node) {
	v.result = n.value
}

func (v *visitor) clear() {
	v.result = nil
}

// was called visit
func (v *visitor) traverse(prefix []byte, n *node) {
	prefixLength := n.getCommonPrefixLength(prefix)

	if prefixLength == len(prefix) && prefixLength == len(n.key) {
		v.visit(n)
		return
	}

	// todo missing `if node == root`

	if prefixLength < len(prefix) && prefixLength >= len(n.key) {
		child := n.findChild(prefix[prefixLength])
		if child != nil {
			newKey := prefix[prefixLength:]
			v.traverse(newKey, child)
			return
		}
	}

	v.clear()
}
