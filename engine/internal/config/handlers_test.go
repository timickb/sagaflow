package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseEndpoints(t *testing.T) {
	t.Run("valid endpoints", func(t *testing.T) {
		cfg := &HandlersConfig{
			CallTimeoutRaw: "",
			CallTimeout:    0,
			Endpoints: []string{
				"id:id:5001",
				"warehouse:warehouse:5002",
			},
		}

		assert.NoError(t, cfg.parseEndpoints())

		idConn, ok := cfg.GetHandlerConnection("id")
		assert.True(t, ok)
		assert.NotNil(t, idConn)
		assert.Equal(t, "id:5001", idConn.Target())

		warehouseConn, ok := cfg.GetHandlerConnection("warehouse")
		assert.True(t, ok)
		assert.NotNil(t, warehouseConn)
		assert.Equal(t, "warehouse:5002", warehouseConn.Target())
	})

	t.Run("invalid endpoints", func(t *testing.T) {
		cfg := &HandlersConfig{
			CallTimeoutRaw: "",
			CallTimeout:    0,
			Endpoints: []string{
				"id:5001",
				"too:many:colons:123:456",
			},
		}

		assert.Error(t, cfg.parseEndpoints())
	})
}
