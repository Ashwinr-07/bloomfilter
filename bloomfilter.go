package bloomfilter

import (
	"encoding/binary"
	"hash"
	"math"
	"sync"

	"hash/fnv"
)

type BloomFilter struct {
	mu        sync.RWMutex
	bitSet    []bool
	size      uint
	hashFuncs []hash.Hash64
	count     uint
}

func New(size uint, numHashes int) *BloomFilter {
	bf := &BloomFilter{
		bitSet:    make([]bool, size),
		size:      size,
		hashFuncs: make([]hash.Hash64, numHashes),
		count:     0,
	}

	for i := 0; i < numHashes; i++ {
		bf.hashFuncs[i] = fnv.New64()
	}

	return bf
}

func (bf *BloomFilter) Add(item []byte) {
	bf.mu.Lock()
	defer bf.mu.Unlock()

	for _, h := range bf.hashFuncs {
		h.Reset()
		h.Write(item)
		index := h.Sum64() % uint64(bf.size)
		bf.bitSet[index] = true
	}
	bf.count++
}

func (bf *BloomFilter) Contains(item []byte) bool {
	bf.mu.RLock()
	defer bf.mu.RUnlock()

	for _, h := range bf.hashFuncs {
		h.Reset()
		h.Write(item)
		index := h.Sum64() % uint64(bf.size)
		if !bf.bitSet[index] {
			return false
		}
	}
	return true
}

func (bf *BloomFilter) Count() uint {
	bf.mu.RLock()
	defer bf.mu.RUnlock()
	return bf.count
}

func (bf *BloomFilter) EstimatedFalsePositiveRate() float64 {
	bf.mu.RLock()
	defer bf.mu.RUnlock()
	k := float64(len(bf.hashFuncs))
	n := float64(bf.count)
	m := float64(bf.size)
	return math.Pow(1 - math.Exp(-k*n/m), k)
}

func (bf *BloomFilter) OptimalNumHashes(expectedElements uint) int {
	return int(math.Ceil(float64(bf.size) / float64(expectedElements) * math.Log(2)))
}

func (bf *BloomFilter) Reset() {
	bf.mu.Lock()
	defer bf.mu.Unlock()
	bf.bitSet = make([]bool, bf.size)
	bf.count = 0
}

func (bf *BloomFilter) Union(other *BloomFilter) *BloomFilter {
	if bf.size != other.size || len(bf.hashFuncs) != len(other.hashFuncs) {
		return nil
	}

	bf.mu.Lock()
	other.mu.RLock()
	defer bf.mu.Unlock()
	defer other.mu.RUnlock()

	result := New(bf.size, len(bf.hashFuncs))
	for i := range bf.bitSet {
		result.bitSet[i] = bf.bitSet[i] || other.bitSet[i]
	}
	result.count = bf.count + other.count

	return result
}

func (bf *BloomFilter) Serialize() []byte {
	bf.mu.RLock()
	defer bf.mu.RUnlock()

	serialized := make([]byte, 8+8+bf.size/8+1)
	binary.LittleEndian.PutUint64(serialized[0:8], uint64(bf.size))
	binary.LittleEndian.PutUint64(serialized[8:16], uint64(bf.count))

	for i, bit := range bf.bitSet {
		if bit {
			serialized[16+i/8] |= 1 << (uint(i) % 8)
		}
	}

	return serialized
}

func Deserialize(data []byte) *BloomFilter {
	size := binary.LittleEndian.Uint64(data[0:8])
	count := binary.LittleEndian.Uint64(data[8:16])

	bf := New(uint(size), 1) 
	bf.count = uint(count)

	for i := uint(0); i < bf.size; i++ {
		if data[16+i/8]&(1<<(i%8)) != 0 {
			bf.bitSet[i] = true
		}
	}

	return bf
}