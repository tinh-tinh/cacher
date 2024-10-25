package cacher

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Bytes(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	person := &Person{
		Name: "John",
		Age:  30,
	}

	data, err := toBytes(person)
	require.Nil(t, err)

	valRaw, err := fromBytes[Person](data)
	require.Nil(t, err)

	val, ok := valRaw.(Person)
	require.True(t, ok)
	require.Equal(t, person.Age, val.Age)
	require.Equal(t, person.Name, val.Name)

	_, err = toBytes(nil)
	require.NotNil(t, err)

	_, err = fromBytes[Person]([]byte("test"))
	require.NotNil(t, err)
}

func Test_Gzip(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	person := Person{
		Name: "John",
		Age:  30,
	}

	data, err := toBytes(person)
	require.Nil(t, err)

	val, err := CompressGzip(data)
	require.Nil(t, err)

	valRaw, err := DecompressGzip(val)
	require.Nil(t, err)

	personRaw, err := fromBytes[Person](valRaw)
	require.Nil(t, err)

	newPerson, ok := personRaw.(Person)
	require.True(t, ok)

	require.Equal(t, person.Age, newPerson.Age)
	require.Equal(t, person.Name, newPerson.Name)
}

func Test_Flate(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	person := Person{
		Name: "John",
		Age:  30,
	}

	data, err := toBytes(person)
	require.Nil(t, err)

	val, err := CompressFlate(data)
	require.Nil(t, err)

	valRaw, err := DecompressFlate(val)
	require.Nil(t, err)

	personRaw, err := fromBytes[Person](valRaw)
	require.Nil(t, err)

	newPerson, ok := personRaw.(Person)
	require.True(t, ok)

	require.Equal(t, person.Age, newPerson.Age)
	require.Equal(t, person.Name, newPerson.Name)
}

func Test_Zlib(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	person := Person{
		Name: "John",
		Age:  30,
	}

	data, err := toBytes(person)
	require.Nil(t, err)

	val, err := CompressZlib(data)
	require.Nil(t, err)

	valRaw, err := DecompressZlib(val)
	require.Nil(t, err)

	personRaw, err := fromBytes[Person](valRaw)
	require.Nil(t, err)

	newPerson, ok := personRaw.(Person)
	require.True(t, ok)

	require.Equal(t, person.Age, newPerson.Age)
	require.Equal(t, person.Name, newPerson.Name)
}
