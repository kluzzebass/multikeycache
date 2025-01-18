package multikeycache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMultiKeyCache(t *testing.T) {
	// create a new cache with duplicate secondary key names
	c, err := NewMultiKeyCache[string, string, string, string]([]string{"a", "a", "c"})
	assert.Nil(t, c)
	assert.ErrorAs(t, err, &ErrorSecondaryKeyNameNotUnique[string]{SecondaryKeyName: "a"})

	// create a new cache with unique secondary key names
	c, err = NewMultiKeyCache[string, string, string, string]([]string{"a", "b", "c"})
	// check if the cache is not nil
	assert.NotNil(t, c)
	assert.Nil(t, err)
	// check if the cache has the correct secondary key names
	assert.Equal(t, []string{"a", "b", "c"}, c.secondaryKeyNames)

	// set the first item
	err = c.Set("pk1", "value", "a1", "b1", "c1")
	assert.Nil(t, err)

	// set the second item with a different number of secondary keys
	err = c.Set("pk2", "value", "a2", "b2")
	assert.ErrorAs(t, err, &ErrorSecondaryKeyNumberMismatch{})

	// set the third item, but the secondary keys already exists for a different primary key
	err = c.Set("pk3", "value", "a1", "b1", "c1")
	assert.ErrorAs(t, err, &ErrorWrongSecondaryKey[string, string]{SecondaryKey: "a", ExistingPK: "pk1", NewPK: "pk3"})

	// get the item by primary key
	value, ok := c.Get("pk1")
	assert.True(t, ok)
	assert.Equal(t, "value", value)

	// get a non-existent item
	value, ok = c.Get("pk4")
	assert.False(t, ok)
	assert.Equal(t, "", value)

	// get the item by secondary key
	value, ok, err = c.GetBySecondaryKey("a", "a1")
	assert.True(t, ok)
	assert.Nil(t, err)
	assert.Equal(t, "value", value)

	// get the item by secondary key name that does not exist
	value, ok, err = c.GetBySecondaryKey("d", "d1")
	assert.False(t, ok)
	assert.ErrorAs(t, err, &ErrorUnknownSecondaryKey[string]{SecondaryKeyName: "d"})
	assert.Equal(t, "", value)

	// get the item by a secondary key, but the item does not exist
	value, ok, err = c.GetBySecondaryKey("a", "a6")
	assert.False(t, ok)
	assert.Nil(t, err)
	assert.Equal(t, "", value)

	// delete the item by secondary key
	err = c.DeleteBySecondaryKey("a", "a1")
	assert.Nil(t, err)

	// get the item by secondary key
	value, ok, err = c.GetBySecondaryKey("a", "a1")
	assert.False(t, ok)
	assert.Nil(t, err)
	assert.Equal(t, "", value)

	// inserts a new item
	err = c.Set("pk4", "value", "a4", "b4", "c4")
	assert.Nil(t, err)

	// check that the item exists
	value, ok = c.Get("pk4")
	assert.True(t, ok)
	assert.Equal(t, "value", value)

	// delete the item by primary key
	c.Delete("pk4")

	// check that the item does not exist
	value, ok = c.Get("pk4")
	assert.False(t, ok)
	assert.Equal(t, "", value)

	// insert a new item
	err = c.Set("pk5", "value", "a5", "b5", "c5")
	assert.Nil(t, err)

	// insert a new item
	err = c.Set("pk6", "value", "a6", "b6", "c6")
	assert.Nil(t, err)

	// check the length of the cache
	assert.Equal(t, 2, c.Len())

	// check the secondary key names
	assert.Equal(t, []string{"a", "b", "c"}, c.SecondaryKeyNames())

	// check the secondary keys
	assert.Equal(t, []string{"a5", "a6"}, c.SecondaryKeys("a"))
	assert.Equal(t, []string{"b5", "b6"}, c.SecondaryKeys("b"))
	assert.Equal(t, []string{"c5", "c6"}, c.SecondaryKeys("c"))

	// check the secondary key name to keys map
	assert.Equal(t, map[string]string{"a5": "pk5", "a6": "pk6"}, c.SecondaryKeyNameToKeys("a"))
	assert.Equal(t, map[string]string{"b5": "pk5", "b6": "pk6"}, c.SecondaryKeyNameToKeys("b"))
	assert.Equal(t, map[string]string{"c5": "pk5", "c6": "pk6"}, c.SecondaryKeyNameToKeys("c"))

	// check the get all
	assert.Equal(t, map[string]string{"pk5": "value", "pk6": "value"}, c.GetAll())

	// clear the cache
	c.Clear()
	assert.Equal(t, 0, c.Len())

	// check the length of the cache
	assert.Equal(t, 0, c.Len())
}
