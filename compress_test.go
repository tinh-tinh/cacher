package cacher_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tinh-tinh/cacher"
)

type BigStruct struct {
	ID          int64
	Name        string
	Description string
	Timestamp   string
	Tags        []string
	Data        []byte
	Metadata    map[string]interface{}
	Coordinates struct {
		Latitude  float64
		Longitude float64
	}
	Attributes struct {
		IsActive    bool
		AccessLevel int
		Notes       []string
	}
	Children []struct {
		ChildID   int
		ChildName string
		ChildData []byte
	}
}

func fillValue() BigStruct {
	bigStruct := BigStruct{
		ID:          123456789,
		Name:        "ExampleStruct",
		Description: "This is a large struct for testing purposes.",
		Timestamp:   time.Now().Format("2006-01-01"),
		Tags:        []string{"example", "testing", "golang"},
		Data:        []byte("Random binary data for testing."),
		Metadata: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
			"key3": []string{"nested", "data"},
		},
		Children: []struct {
			ChildID   int
			ChildName string
			ChildData []byte
		}{
			{ChildID: 1, ChildName: "Child1", ChildData: []byte("Child1 data")},
			{ChildID: 2, ChildName: "Child2", ChildData: []byte("Child2 data")},
		},
	}

	// Fill coordinates and attributes
	bigStruct.Coordinates.Latitude = 37.7749
	bigStruct.Coordinates.Longitude = -122.4194
	bigStruct.Attributes.IsActive = true
	bigStruct.Attributes.AccessLevel = 5
	bigStruct.Attributes.Notes = []string{"Note1", "Note2", "Note3"}

	return bigStruct
}

func Test_Gzip(t *testing.T) {
	bigstruct := fillValue()
	data, err := cacher.ToBytes(bigstruct)
	require.Nil(t, err)

	val, err := cacher.CompressGzip(data)
	require.Nil(t, err)

	require.Less(t, len(val), len(data))

	valRaw, err := cacher.DecompressGzip(val)
	require.Nil(t, err)

	personRaw, err := cacher.FromBytes[BigStruct](valRaw)
	require.Nil(t, err)

	newPerson, ok := personRaw.(BigStruct)
	require.True(t, ok)

	require.Equal(t, bigstruct, newPerson)
}

func Test_Flate(t *testing.T) {
	person := fillValue()
	data, err := cacher.ToBytes(person)
	require.Nil(t, err)

	val, err := cacher.CompressFlate(data)
	require.Nil(t, err)

	valRaw, err := cacher.DecompressFlate(val)
	require.Nil(t, err)

	personRaw, err := cacher.FromBytes[BigStruct](valRaw)
	require.Nil(t, err)

	newPerson, ok := personRaw.(BigStruct)
	require.True(t, ok)

	require.Equal(t, newPerson, person)
}

func Test_Zlib(t *testing.T) {
	person := fillValue()

	data, err := cacher.ToBytes(person)
	require.Nil(t, err)

	val, err := cacher.CompressZlib(data)
	require.Nil(t, err)

	valRaw, err := cacher.DecompressZlib(val)
	require.Nil(t, err)

	personRaw, err := cacher.FromBytes[BigStruct](valRaw)
	require.Nil(t, err)

	newPerson, ok := personRaw.(BigStruct)
	require.True(t, ok)

	require.Equal(t, person, newPerson)
}
