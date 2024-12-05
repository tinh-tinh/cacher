package cacher_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tinh-tinh/cacher"
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

	data, err := cacher.ToBytes(person)
	require.Nil(t, err)

	valRaw, err := cacher.FromBytes[Person](data)
	require.Nil(t, err)

	val, ok := valRaw.(Person)
	require.True(t, ok)
	require.Equal(t, person.Age, val.Age)
	require.Equal(t, person.Name, val.Name)

	_, err = cacher.ToBytes(nil)
	require.NotNil(t, err)

	_, err = cacher.FromBytes[Person]([]byte("test"))
	require.NotNil(t, err)
}
