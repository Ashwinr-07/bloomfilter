package bloomfilter

import (
	"testing"
)

func TestBloomFilter(t *testing.T) {
	bf := New(1000, 3)

	// Test Add and Contains
	bf.Add([]byte("test"))
	if !bf.Contains([]byte("test")) {
		t.Error("BloomFilter should contain 'test'")
	}
	if bf.Contains([]byte("not-added")) {
		t.Error("BloomFilter should not contain 'not-added'")
	}

	// Test Count
	if bf.Count() != 1 {
		t.Errorf("Expected count 1, got %d", bf.Count())
	}

	// Test EstimatedFalsePositiveRate
	fpr := bf.EstimatedFalsePositiveRate()
	if fpr < 0 || fpr > 1 {
		t.Errorf("Invalid false positive rate: %f", fpr)
	}

	// Test OptimalNumHashes
	optimal := bf.OptimalNumHashes(100)
	if optimal <= 0 {
		t.Errorf("Invalid optimal number of hashes: %d", optimal)
	}

	// Test Reset
	bf.Reset()
	if bf.Count() != 0 {
		t.Errorf("Expected count 0 after reset, got %d", bf.Count())
	}

	// Test Union
	bf1 := New(1000, 3)
	bf2 := New(1000, 3)
	bf1.Add([]byte("a"))
	bf2.Add([]byte("b"))
	union := bf1.Union(bf2)
	if !union.Contains([]byte("a")) || !union.Contains([]byte("b")) {
		t.Error("Union should contain elements from both filters")
	}

	// Test Serialize and Deserialize
	serialized := bf.Serialize()
	deserialized := Deserialize(serialized)
	if deserialized.size != bf.size || deserialized.count != bf.count {
		t.Error("Deserialized filter doesn't match original")
	}
}

func BenchmarkBloomFilter(b *testing.B) {
	bf := New(100000, 3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.Add([]byte("test"))
		bf.Contains([]byte("test"))
	}
}