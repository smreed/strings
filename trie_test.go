package strings

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestPutAndGet(t *testing.T) {
	t.Parallel()

	trie := NewPatriciaTrie()

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

func TestEntries(t *testing.T) {
	t.Parallel()

	trie := NewPatriciaTrie()
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

func BenchmarkTriePut4(b *testing.B) {
	intRes, boolRes = benchmarkTriePut(4, b)
}

func BenchmarkTriePut8(b *testing.B) {
	intRes, boolRes = benchmarkTriePut(8, b)
}

func BenchmarkTriePut24(b *testing.B) {
	intRes, boolRes = benchmarkTriePut(24, b)
}

func BenchmarkTriePut256(b *testing.B) {
	intRes, boolRes = benchmarkTriePut(256, b)
}

func BenchmarkTriePut1024(b *testing.B) {
	intRes, boolRes = benchmarkTriePut(1024, b)
}

func benchmarkTriePut(size int, b *testing.B) (int, bool) {
	b.ReportAllocs()
	t := NewPatriciaTrie()
	for n := 0; n < b.N; n++ {
		t.Put(randomString(size), struct{}{})
	}
	fmt.Printf("size: %d, b.N=%d, size*b.N=%d, est=%d, ratio=%f\n", size, b.N, size*b.N, t.estimateSize(), float64(t.estimateSize())/float64(b.N*size))
	return t.Size, t.Contains(randomString(size))
}

// this currently doesn't benchmark well because it loads a TON of matches into memory
func BenchmarkTrieScan1x4x1024(b *testing.B) {
	intRes, boolRes = benchmarkTrieScan(1, 4, 1024, b)
}

func BenchmarkTrieScan4x4x1024(b *testing.B) {
	intRes, boolRes = benchmarkTrieScan(4, 4, 1024, b)
}

func BenchmarkTrieScan256x4x1024(b *testing.B) {
	intRes, boolRes = benchmarkTrieScan(256, 4, 1024, b)
}

// this currently doesn't benchmark well because it loads a TON of matches into memory
func BenchmarkTrieScan1x8x1024(b *testing.B) {
	intRes, boolRes = benchmarkTrieScan(1, 8, 1024, b)
}

func BenchmarkTrieScan4x8x1024(b *testing.B) {
	intRes, boolRes = benchmarkTrieScan(4, 8, 1024, b)
}

func BenchmarkTrieScan256x8x1024(b *testing.B) {
	intRes, boolRes = benchmarkTrieScan(256, 8, 1024, b)
}

func benchmarkTrieScan(psize, size, n int, b *testing.B) (int, bool) {
	t := NewPatriciaTrie()
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
