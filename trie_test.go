package strings

import (
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/armon/go-radix"
	"github.com/hashicorp/go-immutable-radix"
)

func TestPutAndGet(t *testing.T) {
	t.Parallel()

	trie := NewTrie()

	m := map[string]interface{}{
		"foo":           "foovalue",
		"bar":           "barvalue",
		"foobar":        "foobarvalue",
		"baloney":       42,
		randomString(8): randomString(1024),
		randomString(8): randomString(1024),
		randomString(8): randomString(1024),
		randomString(8): randomString(1024),
		randomString(8): randomString(1024),
		randomString(8): randomString(1024),
		randomString(8): randomString(1024),
		randomString(8): randomString(1024),
	}

	for k, v := range m {
		trie.Put(k, v)
	}

	for k, v := range m {
		actual := trie.Get(k)
		if v != actual {
			t.Errorf("Expect value of %v for %v, got %v", v, k, actual)
		}
	}
}

func TestContainsPrefix(t *testing.T) {
	trie := NewTrie()
	trie.Put("foo", "foovalue")
	trie.Put("bar", "barvalue")
	trie.Put("foobar", "foobarvalue")
	trie.Put("f", "F")

	m := map[string]bool{
		"f":       true,
		"fo":      true,
		"foo":     true,
		"foob":    true,
		"fooba":   true,
		"foobar":  true,
		"foobarf": false,
		"b":       true,
		"ba":      true,
		"bar":     true,
		"barf":    false,
		"c":       false,
	}

	for k, v := range m {
		actual := trie.ContainsPrefix(k)
		if v != actual {
			t.Errorf("Expect value of %v for %v, got %v", v, k, actual)
		}
	}
}

func TestEntries(t *testing.T) {
	t.Parallel()

	trie := NewTrie()
	// Put in an explicit order to test node rewrites
	trie.Put("foobar", "foobarvalue")
	trie.Put("foo", "foovalue")
	trie.Put("baloney", 42)
	trie.Put("bar", "barvalue")

	m := map[string]interface{}{
		"foo":     "foovalue",
		"baloney": 42,
		"bar":     "barvalue",
		"foobar":  "foobarvalue",
	}

	mm := trie.Entries()

	if !reflect.DeepEqual(m, mm) {
		t.Errorf("Expect map %v, got %v", m, mm)
	}
}

func TestGetCommonPrefixLength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		left, right string
		expected    int
	}{
		{left: "", right: "bcd", expected: 0},
		{left: "abc", right: "", expected: 0},
		{left: "abc", right: "bcd", expected: 0},
		{left: "abc", right: "ABC", expected: 0},
		{left: "abc", right: "acd", expected: 1},
		{left: "abc", right: "abd", expected: 2},
		{left: "abc", right: "abc", expected: 3},
		{left: "abc", right: "abcd", expected: 3},
	}

	for _, test := range tests {
		n := &node{key: []byte(test.left)}
		actual := n.getCommonPrefixLength([]byte(test.right))
		if actual != test.expected {
			t.Errorf("Expect common prefix length of %q and %q to be %d, got %d", test.left, test.right, test.expected, actual)
		}
	}
}

func TestCompare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		left, right string
		expected    int
	}{
		{left: "", right: "abc", expected: -3},
		{left: "abc", right: "", expected: 3},
		{left: "abc", right: "bcd", expected: -1},
		{left: "abc", right: "ABC", expected: 1},
		{left: "abc", right: "acd", expected: -1},
		{left: "abc", right: "abd", expected: -1},
		{left: "abc", right: "abc", expected: 0},
		{left: "abc", right: "abcd", expected: -1},
	}

	for _, test := range tests {
		n := &node{key: []byte(test.left)}
		actual := n.compareTo(&node{key: []byte(test.right)})
		if actual != test.expected {
			t.Errorf("Expect compare of %q and %q to be %d, got %d", test.left, test.right, test.expected, actual)
		}
	}
}

func TestKeyStartsWith(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key      string
		b        byte
		expected bool
	}{
		{key: "", b: 'a', expected: false},
		{key: "abc", b: 'b', expected: false},
		{key: "abc", b: 'A', expected: false},
		{key: "abc", b: 'a', expected: true},
	}

	for _, test := range tests {
		n := &node{key: []byte(test.key)}
		actual := n.keyStartsWith(test.b)
		if actual != test.expected {
			t.Errorf("Expect keyStartsWith of %q and %q to be %v, got %v", test.key, test.b, test.expected, actual)
		}
	}
}

var intRes int
var boolRes bool

func BenchmarkTriePut24(b *testing.B) {
	intRes, boolRes = benchmarkTriePut(NewTrie(), 24, b)
}

func BenchmarkGoRadixTriePut24(b *testing.B) {
	intRes, boolRes = benchmarkTriePut(goradixTrie(), 24, b)
}

func BenchmarkImmutableTriePut24(b *testing.B) {
	intRes, boolRes = benchmarkTriePut(hashicorpImmutableTrie(), 24, b)
}

func BenchmarkTriePut256(b *testing.B) {
	intRes, boolRes = benchmarkTriePut(NewTrie(), 256, b)
}

func BenchmarkGoRadixTriePut256(b *testing.B) {
	intRes, boolRes = benchmarkTriePut(goradixTrie(), 256, b)
}

func BenchmarkImmutableTriePut256(b *testing.B) {
	intRes, boolRes = benchmarkTriePut(hashicorpImmutableTrie(), 256, b)
}

func BenchmarkTriePut1024(b *testing.B) {
	intRes, boolRes = benchmarkTriePut(NewTrie(), 1024, b)
}

func BenchmarkGoRadixTriePut1024(b *testing.B) {
	intRes, boolRes = benchmarkTriePut(goradixTrie(), 1024, b)
}

func BenchmarkImmutableTriePut1024(b *testing.B) {
	intRes, boolRes = benchmarkTriePut(hashicorpImmutableTrie(), 1024, b)
}

func goradixTrie() trie {
	t := radix.New()
	return &goradixAdapter{t: t}
}

type goradixAdapter struct {
	t *radix.Tree
}

func (t *goradixAdapter) Put(k string, v interface{}) {
	t.t.Insert(k, v)
}

func (t *goradixAdapter) Contains(k string) bool {
	_, exists := t.t.Get(k)
	return exists
}

func (t *goradixAdapter) ContainsPrefix(k string) bool {
	prefix, _, _ := t.t.LongestPrefix(k)
	return prefix == k
}

func (t *goradixAdapter) Size() int {
	return t.t.Len()
}

func hashicorpImmutableTrie() trie {
	t := iradix.New()
	return &iradixAdapter{t: t}
}

type iradixAdapter struct {
	t *iradix.Tree
}

func (t *iradixAdapter) Put(k string, v interface{}) {
	t.t, _, _ = t.t.Insert([]byte(k), v)
}

func (t *iradixAdapter) Contains(k string) bool {
	_, exists := t.t.Get([]byte(k))
	return exists
}

func (t *iradixAdapter) ContainsPrefix(k string) bool {
	prefix, _, _ := t.t.Root().LongestPrefix([]byte(k))
	return string(prefix) == k
}

func (t *iradixAdapter) Size() int {
	return t.t.Len()
}

type trie interface {
	Put(k string, v interface{})
	Contains(k string) bool
	ContainsPrefix(k string) bool
	Size() int
}

func benchmarkTriePut(t trie, size int, b *testing.B) (int, bool) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		t.Put(randomString(size), struct{}{})
	}
	return t.Size(), t.Contains(randomString(size))
}

// this currently doesn't benchmark well because it loads a TON of matches into memory
func BenchmarkTrieScan1x4x1024(b *testing.B) {
	intRes, boolRes = benchmarkTrieScan(NewTrie(), 1, 4, 1024, b)
}

func BenchmarkGoRadixTrieScan1x4x1024(b *testing.B) {
	intRes, boolRes = benchmarkTrieScan(goradixTrie(), 1, 4, 1024, b)
}

func BenchmarkTrieScan4x4x1024(b *testing.B) {
	intRes, boolRes = benchmarkTrieScan(NewTrie(), 4, 4, 1024, b)
}

func BenchmarkGoRadixTrieScan4x4x1024(b *testing.B) {
	intRes, boolRes = benchmarkTrieScan(goradixTrie(), 4, 4, 1024, b)
}

func BenchmarkTrieScan256x4x1024(b *testing.B) {
	intRes, boolRes = benchmarkTrieScan(NewTrie(), 256, 4, 1024, b)
}

func BenchmarkGoRadixTrieScan256x4x1024(b *testing.B) {
	intRes, boolRes = benchmarkTrieScan(goradixTrie(), 256, 4, 1024, b)
}

// this currently doesn't benchmark well because it loads a TON of matches into memory
func BenchmarkTrieScan1x8x1024(b *testing.B) {
	intRes, boolRes = benchmarkTrieScan(NewTrie(), 1, 8, 1024, b)
}

func BenchmarkGoRadixTrieScan1x8x1024(b *testing.B) {
	intRes, boolRes = benchmarkTrieScan(goradixTrie(), 1, 8, 1024, b)
}

func BenchmarkTrieScan4x8x1024(b *testing.B) {
	intRes, boolRes = benchmarkTrieScan(NewTrie(), 4, 8, 1024, b)
}

func BenchmarkGoRadixTrieScan4x8x1024(b *testing.B) {
	intRes, boolRes = benchmarkTrieScan(goradixTrie(), 4, 8, 1024, b)
}

func BenchmarkTrieScan256x8x1024(b *testing.B) {
	intRes, boolRes = benchmarkTrieScan(NewTrie(), 256, 8, 1024, b)
}

func BenchmarkGoRadixTrieScan256x8x1024(b *testing.B) {
	intRes, boolRes = benchmarkTrieScan(goradixTrie(), 256, 8, 1024, b)
}

func benchmarkTrieScan(t trie, psize, size, n int, b *testing.B) (int, bool) {
	for i := 0; i < n; i++ {
		t.Put(randomString(size), struct{}{})
	}
	matches := 0
	prefix := randomString(psize)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if bb := t.ContainsPrefix(prefix); bb {
			matches++
		}
	}
	return matches, matches > 0
}

// Random string generation from
// http://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func randomString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
