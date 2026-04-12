package dsl

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/lib/utils"
)

func TestCache_NewCache(t *testing.T) {
	cache, err := NewCache("./testdata")
	require.NoError(t, err)
	require.NotNil(t, cache)

	expectedHeader1 := domain.SagaDefinitionHeader{
		Name:    "Process Order",
		Version: 1,
	}
	expectedHeader2 := domain.SagaDefinitionHeader{
		Name:    "Process Order",
		Version: 1,
	}

	require.Contains(
		t,
		utils.MapToKeysSlice(cache.definitions),
		expectedHeader1,
	)
	require.Contains(
		t,
		utils.MapToKeysSlice(cache.definitions),
		expectedHeader2,
	)

	sagaDef1 := cache.definitions[expectedHeader1]
	require.NotNil(t, sagaDef1)
	require.Equal(t, expectedHeader1.Name, sagaDef1.Name)
	require.Equal(t, expectedHeader1.Version, sagaDef1.Version)

	sagaDef2 := cache.definitions[expectedHeader2]
	require.NotNil(t, sagaDef1)
	require.Equal(t, expectedHeader2.Name, sagaDef2.Name)
	require.Equal(t, expectedHeader2.Version, sagaDef2.Version)
}
