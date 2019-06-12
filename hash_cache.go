package ssz

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/karlseguin/ccache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// ErrNotMerkleRoot will be returned when a cache object is not a merkle root.
	ErrNotMerkleRoot = errors.New("object is not a merkle root")
	// Metrics
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

// hashCacheS struct with one queue for looking up by hash.
type hashCacheS struct {
	hashCache *ccache.Cache
}

// root specifies the hash of data in a struct
type root struct {
	Hash       common.Hash
	MerkleRoot []byte
}

// newHashCache creates a new hash cache for storing/accessing root hashes from
// memory.
func newHashCache(maxCacheSize int64) *hashCacheS {
	return &hashCacheS{
		hashCache: ccache.New(ccache.Configure().MaxSize(maxCacheSize)),
	}
}

// RootByEncodedHash fetches Root by the encoded hash of the object. Returns true with a
// reference to the root if exists. Otherwise returns false, nil.
func (b *hashCacheS) RootByEncodedHash(h common.Hash) (bool, *root, error) {
	item := b.hashCache.Get(h.Hex())
	if item == nil {
		hashCacheMiss.Inc()
		return false, nil, nil
	}
	hashCacheHit.Inc()
	hInfo, ok := item.Value().(*root)
	if !ok {
		return false, nil, ErrNotMerkleRoot
	}

	return true, hInfo, nil
}

func (b *hashCacheS) lookup(rval reflect.Value, hasher hasher) ([32]byte, error) {
	hs, err := HashedEncoding(rval)
	if err != nil {
		return [32]byte{}, newHashError(fmt.Sprint(err), rval.Type())
	}
	exists, fetchedInfo, err := b.RootByEncodedHash(hs)
	if err != nil {
		return [32]byte{}, newHashError(fmt.Sprint(err), rval.Type())
	}
	if exists {
		return ToBytes32(fetchedInfo.MerkleRoot), nil
	}
	res, err := hasher(rval)
	if err != nil {
		return [32]byte{}, err
	}
	err = b.AddRoot(hs, res[:])
	if err != nil {
		return [32]byte{}, err
	}
	return hasher(rval)

}

// AddRoot adds an encodedhash of the object as key and a rootHash object to the cache.
// This method also trims the
// least recently added root info if the cache size has reached the max cache
// size limit.
func (b *hashCacheS) AddRoot(h common.Hash, rootB []byte) error {
	mr := &root{
		Hash:       h,
		MerkleRoot: rootB,
	}
	b.hashCache.Set(mr.Hash.Hex(), mr, time.Hour)
	return nil
}
