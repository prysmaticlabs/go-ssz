package ssz

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// The marshaler/unmarshaler types take in a value, an output buffer, and a start offset,
// it returns the index of the last byte written and an error, if any.
type marshaler func(reflect.Value, []byte, uint64) (uint64, error)

type unmarshaler func([]byte, reflect.Value, uint64) (uint64, error)

type hasher func(reflect.Value) ([32]byte, error)

type sszUtils struct {
	marshaler
	unmarshaler
	hasher
}

var (
	sszUtilsCacheMutex sync.RWMutex
	sszUtilsCache      = make(map[reflect.Type]*sszUtils)
	hashCache          = newHashCache(100000)
)

// Get cached encoder, encodeSizer and unmarshaler implementation for a specified type.
// With a cache we can achieve O(1) amortized time overhead for creating encoder, encodeSizer and decoder.
func cachedSSZUtils(typ reflect.Type) (*sszUtils, error) {
	sszUtilsCacheMutex.RLock()
	utils := sszUtilsCache[typ]
	sszUtilsCacheMutex.RUnlock()
	if utils != nil {
		return utils, nil
	}

	// If not found in cache, will get a new one and put it into the cache
	sszUtilsCacheMutex.Lock()
	defer sszUtilsCacheMutex.Unlock()
	return cachedSSZUtilsNoAcquireLock(typ)
}

// This version is used when the caller is already holding the rw lock for sszUtilsCache.
// It doesn't acquire new rw lock so it's free to recursively call itself without getting into
// a deadlock situation.
//
// Make sure you are
func cachedSSZUtilsNoAcquireLock(typ reflect.Type) (*sszUtils, error) {
	// Check again in case other goroutine has just acquired the lock
	// and already updated the cache
	utils := sszUtilsCache[typ]
	if utils != nil {
		return utils, nil
	}
	// Put a dummy value into the cache before generating.
	// If the generator tries to lookup the type of itself,
	// it will get the dummy value and won't call recursively forever.
	sszUtilsCache[typ] = new(sszUtils)
	utils, err := generateSSZUtilsForType(typ)
	if err != nil {
		// Don't forget to remove the dummy key when fail
		delete(sszUtilsCache, typ)
		return nil, err
	}
	// Overwrite the dummy value with real value
	*sszUtilsCache[typ] = *utils
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

type field struct {
	index    int
	name     string
	typ      reflect.Type
	sszUtils *sszUtils
}

// truncateLast removes the last value of a struct, usually the signature,
// in order to hash only the data the signature field is intended to represent.
func truncateLast(typ reflect.Type) (fields []field, err error) {
	fields, err = marshalerStructFields(typ)
	if err != nil {
		return nil, err
	}
	return fields[:len(fields)-1], nil
}

func marshalerStructFields(typ reflect.Type) (fields []field, err error) {
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if strings.Contains(f.Name, "XXX") {
			continue
		}
		fType, err := fieldType(f)
		if err != nil {
			return nil, err
		}
		utils, err := cachedSSZUtilsNoAcquireLock(fType)
		if err != nil {
			return nil, fmt.Errorf("failed to get ssz utils: %v", err)
		}
		name := f.Name
		fields = append(fields, field{index: i, name: name, sszUtils: utils, typ: fType})
	}
	return fields, nil
}

func unmarshalerStructFields(typ reflect.Type) (fields []field, err error) {
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if strings.Contains(f.Name, "XXX") {
			continue
		}
		fType, err := fieldType(f)
		if err != nil {
			return nil, err
		}
		utils, err := cachedSSZUtilsNoAcquireLock(f.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to get ssz utils: %v", err)
		}
		name := f.Name
		fields = append(fields, field{index: i, name: name, sszUtils: utils, typ: fType})
	}
	return fields, nil
}

func sszTagSizes(tag string) ([]int, error) {
	sizeStartIndex := strings.IndexRune(tag, '=')
	items := strings.Split(tag[sizeStartIndex+1:], ",")
	sizes := make([]int, len(items))
	var err error
	for i := 0; i < len(items); i++ {
		if items[i] == "?" {
			// TODO: find a way to handle this nicely.
			sizes[i] = 0
			continue
		}
		sizes[i], err = strconv.Atoi(items[i])
		if err != nil {
			return nil, err
		}
	}
	return sizes, nil
}

func fieldType(field reflect.StructField) (reflect.Type, error) {
	item, exists := field.Tag.Lookup("ssz")
	if exists {
		sizes, err := sszTagSizes(item)
		if err != nil {
			return nil, err
		}
		if field.Type.Elem().Kind() == reflect.Slice {
			if len(sizes) > 1 {
				innerData := reflect.ArrayOf(sizes[1], field.Type.Elem().Elem())
				if sizes[0] == 0 {
					return reflect.SliceOf(innerData), nil
				}
				return reflect.ArrayOf(sizes[0], innerData), nil
			} else {
				innerData := reflect.ArrayOf(sizes[0], field.Type.Elem().Elem())
				return reflect.SliceOf(innerData), nil
			}
		}
		return reflect.ArrayOf(sizes[0], field.Type.Elem()), nil
	}
	return field.Type, nil
}
