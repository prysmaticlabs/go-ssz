package ssz

import (
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	BytesPerChunk = 32 // BytesPerChunk for an SSZ serialized object.
	BytesPerLengthOffset = 4 // BytesPerLengthOffset defines a constant for off-setting serialized chunks.
	BitsPerByte = 8 // BitsPerByte as a useful constant.
	ErrNotMerkleRoot = errors.New("object is not a Merkle root") // ErrNotMerkleRoot when a cached item is not root.
	hashCacheMiss = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ssz_hash_cache_miss",
		Help: "The number of hash requests that aren't present in the cache.",
	})
	hashCacheHit = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ssz_hash_cache_hit",
		Help: "The number of hash requests that are present in the cache.",
	})
	hashCacheSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ssz_hash_cache_size",
		Help: "The number of hashes in the block cache",
	})
)
