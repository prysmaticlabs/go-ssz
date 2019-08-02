package ssz

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/karlseguin/ccache"
	"github.com/minio/highwayhash"
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
	Hash       []byte
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
func (b *hashCacheS) RootByEncodedHash(h []byte) (bool, *root, error) {
	item := b.hashCache.Get(string(h))
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

func (b *hashCacheS) lookup(
	rval reflect.Value,
	hasher hasher,
	marshaler marshaler,
	maxCapacity uint64,
) ([32]byte, error) {
	cacheKey, err := generateCacheKey(rval, marshaler, maxCapacity)
	if err != nil {
		return [32]byte{}, err
	}
	// We take the hash of the generated cache key.
	h, _ := highwayhash.New(make([]byte, 32))

	if _, err := h.Write(cacheKey); err != nil {
		return [32]byte{}, err
	}
	hs := h.Sum(nil)
	exists, fetchedInfo, err := b.RootByEncodedHash(hs)
	if err != nil {
		return [32]byte{}, err
	}
	if exists {
		return toBytes32(fetchedInfo.MerkleRoot), nil
	}
	res, err := hasher(rval, maxCapacity)
	if err != nil {
		return [32]byte{}, err
	}
	err = b.AddRoot(hs, res[:])
	if err != nil {
		return [32]byte{}, err
	}
	return res, nil
}

// AddRoot adds an encodedhash of the object as key and a rootHash object to the cache.
// This method also trims the
// least recently added root info if the cache size has reached the max cache
// size limit.
func (b *hashCacheS) AddRoot(h []byte, rootB []byte) error {
	mr := &root{
		Hash:       h,
		MerkleRoot: rootB,
	}
	b.hashCache.Set(string(h), mr, time.Hour)
	hashCacheSize.Set(float64(b.hashCache.ItemCount()))
	return nil
}

func generateCacheKey(v reflect.Value, marshaler marshaler, maxCapacity uint64) ([]byte, error) {
	encodedLength := make([]byte, 8)
	encodedCapacity := make([]byte, 8)
	binary.LittleEndian.PutUint64(encodedCapacity, maxCapacity)
	var buf []byte
	var err error
	if v.Kind() == reflect.Struct {
		buf, err = generateStructHashKey(v)
		if err != nil {
			return nil, err
		}
	} else {
		if v.Kind() != reflect.Struct || (v.Kind() == reflect.Ptr && !v.IsNil()) {
			buf = make([]byte, determineSize(v))
			if _, err := marshaler(v, buf, 0); err != nil {
				return nil, err
			}
			binary.LittleEndian.PutUint64(encodedLength, uint64(len(buf)))
		}
		buf = append(buf, []byte(v.Type().String())...)
	}
	lengthMetadata := append(encodedCapacity, encodedLength...)
	buf = append(buf, lengthMetadata...)
	return buf, nil
}

func generateStructHashKey(v reflect.Value) ([]byte, error) {
	t := v.Type()
	fields, err := structFields(t)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteString(t.String())
	for i := 0; i < len(fields); i++ {
		f := fields[i]
		buf.WriteString(f.typ.String())
		buf.WriteString(f.name)
		if f.typ.Kind() == reflect.Array {
			buf.WriteString(fmt.Sprintf("%d", f.typ.Len()))
		}
		if f.typ.Kind() == reflect.Slice {
			buf.WriteString(fmt.Sprintf("%d", v.Field(f.index).Len()))
		}
		buf.WriteString(fmt.Sprintf("%d", f.capacity))
		buf.WriteString(fmt.Sprintf("%v", v.Field(f.index).Interface()))
	}
	buf.WriteString(string(len(fields)))
	return buf.Bytes(), nil
}
