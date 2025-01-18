package multikeycache

import (
	"fmt"
	"sync"
)

// To avoid confusing myself with the generic types, I'm using the following naming conventions:
// - PKT: (PrimaryKeyType) The type of the primary key
// - VT: (ValueType) The type of the value stored in the cache
// - SKNT: (SecondaryKeyNameType) The type of the secondary key name
// - SKT: (SecondaryKeyType) The type of the secondary key

type ErrorSecondaryKeyNumberMismatch struct {
	Expected int
	Actual   int
}

func (e ErrorSecondaryKeyNumberMismatch) Error() string {
	return fmt.Sprintf("number of secondary keys does not match number of secondary key names: expected %d, actual %d", e.Expected, e.Actual)
}

type ErrorWrongSecondaryKey[PKT comparable, SKNT comparable] struct {
	ExistingPK   PKT
	NewPK        PKT
	SecondaryKey SKNT
}

func (e ErrorWrongSecondaryKey[PKT, SKNT]) Error() string {
	return fmt.Sprintf("secondary key %v already exists for a different pk %v", e.SecondaryKey, e.ExistingPK)
}

type ErrorUnknownSecondaryKey[SKNT comparable] struct {
	SecondaryKeyName SKNT
}

func (e ErrorUnknownSecondaryKey[SKNT]) Error() string {
	return fmt.Sprintf("secondary key not found for secondary key name %v", e.SecondaryKeyName)
}

type ErrorSecondaryKeyNameNotUnique[SKNT comparable] struct {
	SecondaryKeyName SKNT
}

func (e ErrorSecondaryKeyNameNotUnique[SKNT]) Error() string {
	return fmt.Sprintf("secondary key name %v is not unique", e.SecondaryKeyName)
}

type item[PKT comparable, VT any, SecondaryKeyNameType comparable, SKT comparable] struct {
	pk            PKT
	value         VT
	secondaryKeys map[SecondaryKeyNameType]SKT
}

type multiKeyCache[PKT comparable, VT any, SKNT comparable, SKT comparable] struct {
	mu                sync.RWMutex
	values            map[PKT]item[PKT, VT, SKNT, SKT]
	indexes           map[SKNT]map[SKT]PKT
	secondaryKeyNames []SKNT
}

// NewMultiKeyCache creates a new multi-key cache
// and returns an error if the secondary key names are not unique
func NewMultiKeyCache[PKT comparable, VT any, SKNT comparable, SKT comparable](secondaryKeyNames []SKNT) (*multiKeyCache[PKT, VT, SKNT, SKT], error) {
	c := &multiKeyCache[PKT, VT, SKNT, SKT]{
		values:            make(map[PKT]item[PKT, VT, SKNT, SKT]),
		indexes:           make(map[SKNT]map[SKT]PKT),
		secondaryKeyNames: make([]SKNT, len(secondaryKeyNames)),
	}

	// check if the secondary key names are unique
	seen := make(map[SKNT]bool)
	for _, name := range secondaryKeyNames {
		if seen[name] {
			return nil, ErrorSecondaryKeyNameNotUnique[SKNT]{SecondaryKeyName: name}
		}
		seen[name] = true
	}

	for i, name := range secondaryKeyNames {
		c.secondaryKeyNames[i] = name
		c.indexes[name] = make(map[SKT]PKT)
	}

	return c, nil
}

// Set sets the value of the item with the given primary key
// and the given secondary keys (in the same order as the secondary key names)
// and returns an error if the secondary keys do not match the secondary key names
func (c *multiKeyCache[PKT, VT, SKNT, SKT]) Set(pk PKT, v VT, sKeys ...SKT) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// check if the number of secondary keys matches the number of secondary key names
	if len(sKeys) != len(c.secondaryKeyNames) {
		return ErrorSecondaryKeyNumberMismatch{Expected: len(c.secondaryKeyNames), Actual: len(sKeys)}
	}

	// check if the secondary keys already exist for a different pk
	for i, k := range c.secondaryKeyNames {
		if spk, ok := c.indexes[k][sKeys[i]]; ok {
			if spk != pk {
				return ErrorWrongSecondaryKey[PKT, SKNT]{SecondaryKey: k, ExistingPK: spk, NewPK: pk}
			}
		}
	}

	// create the item
	item := item[PKT, VT, SKNT, SKT]{
		pk:            pk,
		value:         v,
		secondaryKeys: make(map[SKNT]SKT),
	}

	// set the secondary keys
	for i, sKey := range sKeys {
		item.secondaryKeys[c.secondaryKeyNames[i]] = sKey
	}

	// set the item in the cache
	c.values[pk] = item

	// set the secondary keys in the indexes
	for _, k := range c.secondaryKeyNames {
		c.indexes[k][item.secondaryKeys[k]] = pk
	}

	return nil
}

// Get returns the value of the item with the given primary key
// and a boolean indicating if the item was found
func (c *multiKeyCache[PKT, VT, SKNT, SKT]) Get(pk PKT) (VT, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// get the item by primary key
	item, ok := c.values[pk]
	if !ok {
		var v VT
		return v, false
	}

	return item.value, true
}

// GetBySecondaryKey returns the value of the item with the given secondary key
// and a boolean indicating if the item was found
// and an error if the secondary key name does not exist
func (c *multiKeyCache[PKT, VT, SKNT, SKT]) GetBySecondaryKey(skn SKNT, sk SKT) (VT, bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var zero VT

	// check if the secondary key name exists
	if !c.secondaryKeyNameExists(skn) {
		return zero, false, ErrorUnknownSecondaryKey[SKNT]{SecondaryKeyName: skn}
	}

	// check if the secondary key exists
	pk, ok := c.indexes[skn][sk]
	if !ok {
		return zero, false, nil
	}

	// get the item by primary key
	value, ok := c.Get(pk)

	return value, ok, nil
}

// Delete deletes the item with the given primary key
func (c *multiKeyCache[PKT, VT, SKNT, SKT]) Delete(pk PKT) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// find the item
	item, ok := c.values[pk]
	if !ok {
		return
	}

	// delete the item
	delete(c.values, pk)

	// delete the secondary keys from the indexes
	for _, skn := range c.secondaryKeyNames {
		delete(c.indexes[skn], item.secondaryKeys[skn])
	}
}

// DeleteBySecondaryKey deletes the item with the given secondary key
// and returns an error if the secondary key name does not exist
func (c *multiKeyCache[PKT, VT, SKNT, SKT]) DeleteBySecondaryKey(skn SKNT, sk SKT) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// check if the secondary key name exists
	if !c.secondaryKeyNameExists(skn) {
		return ErrorUnknownSecondaryKey[SKNT]{SecondaryKeyName: skn}
	}

	// find the item
	pk, ok := c.indexes[skn][sk]
	if !ok {
		return nil
	}

	// find the item
	item, ok := c.values[pk]
	if !ok {
		return nil
	}

	// delete the item by primary key
	delete(c.values, pk)

	// delete the secondary keys from the indexes
	for _, skn := range c.secondaryKeyNames {
		delete(c.indexes[skn], item.secondaryKeys[skn])
	}

	return nil
}

// secondaryKeyNameExists returns true if the secondary key name exists
// and false otherwise
func (c *multiKeyCache[PKT, VT, SKNT, SKT]) secondaryKeyNameExists(skn SKNT) bool {
	for _, n := range c.secondaryKeyNames {
		if n == skn {
			return true
		}
	}

	return false
}

// Clear clears the entire cache. All of it. Gone.
func (c *multiKeyCache[PKT, VT, SKNT, SKT]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.values = make(map[PKT]item[PKT, VT, SKNT, SKT])
	c.indexes = make(map[SKNT]map[SKT]PKT)
}

// Len returns the number of items in the cache
func (c *multiKeyCache[PKT, VT, SKNT, SKT]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.values)
}
