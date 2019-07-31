package ssz

import (
	"reflect"
	"sync"
)

// The marshaler/unmarshaler types take in a value, an output buffer, and a start offset,
// it returns the index of the last byte written and an error, if any.
type marshaler func(reflect.Value, []byte, uint64) (uint64, error)

type unmarshaler func([]byte, reflect.Value, uint64) (uint64, error)

type hasher func(reflect.Value, uint64) ([32]byte, error)

type sszUtils struct {
	marshaler
	unmarshaler
	hasher
}

var (
	sszUtilsCacheMutex sync.Mutex
	sszUtilsCache      = make(map[reflect.Type]*sszUtils)
	hashCache          = newHashCache(100000)
)

// Get cached encoder, encodeSizer and unmarshaler implementation for a specified type.
// With a cache we can achieve O(1) amortized time overhead for creating encoder, encodeSizer and decoder.
func cachedSSZUtils(typ reflect.Type) (*sszUtils, error) {
	sszUtilsCacheMutex.Lock()
	cachedUtils, ok := sszUtilsCache[typ]
	sszUtilsCacheMutex.Unlock()
	if ok && cachedUtils != nil {
		return cachedUtils, nil
	}
	// Put a dummy value into the cache before generating.
	// If the generator tries to lookup the type of itself,
	// it will get the dummy value and won't call recursively forever.
	sszUtilsCacheMutex.Lock()
	sszUtilsCache[typ] = new(sszUtils)
	sszUtilsCacheMutex.Unlock()
	utils, err := generateSSZUtilsForType(typ)
	if err != nil {
		sszUtilsCacheMutex.Lock()
		delete(sszUtilsCache, typ)
		sszUtilsCacheMutex.Unlock()
		return nil, err
	}
	// Overwrite the dummy value with real value.
	sszUtilsCacheMutex.Lock()
	*sszUtilsCache[typ] = *utils
	sszUtilsCacheMutex.Unlock()
	return sszUtilsCache[typ], nil
}

func generateSSZUtilsForType(typ reflect.Type) (utils *sszUtils, err error) {
	utils = new(sszUtils)
	if utils.marshaler, err = makeMarshaler(typ); err != nil {
		return nil, err
	}
	if utils.unmarshaler, err = makeUnmarshaler(typ); err != nil {
		return nil, err
	}
	if utils.hasher, err = makeHasher(typ); err != nil {
		return nil, err
	}
	return utils, nil
}
