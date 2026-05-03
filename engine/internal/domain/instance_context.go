package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

var (
	EmptyInstanceContext = &JsonInstanceContext{
		raw:  []byte("{}"),
		data: make(map[string]any),
	}
)

type InstanceContext interface {
	GetRaw() json.RawMessage
	AppendMap(data map[string]any) (InstanceContext, error)
	Find(path string) (any, error)
}

type JsonInstanceContext struct {
	raw  json.RawMessage
	data any
}

func NewJsonInstanceContextFromRaw(raw json.RawMessage) (*JsonInstanceContext, error) {
	var parsed any
	if len(raw) > 0 && string(raw) != "null" {
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.UseNumber()

		if err := dec.Decode(&parsed); err != nil {
			return nil, fmt.Errorf("decode raw json: %w", err)
		}
	}

	return &JsonInstanceContext{
		raw:  raw,
		data: parsed,
	}, nil
}

func NewJsonInstanceContextFromAny(data any) (*JsonInstanceContext, error) {
	parsed, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal map: %w", err)
	}
	return &JsonInstanceContext{raw: parsed, data: data}, nil
}

func NewStepInputContext(stepInputs []StepInputParam, initialCtx, runtimeCtx InstanceContext) (*JsonInstanceContext, error) {
	data := make(map[string]any)
	for _, inputData := range stepInputs {
		switch inputData.SourceNamespace {
		case StepInputSourceInputContext:
			value, findErr := initialCtx.Find(inputData.SourcePath)
			if findErr != nil {
				return nil, fmt.Errorf("find step input value in initial context: %w", findErr)
			}
			data[inputData.DestinationParam] = value
		case StepInputSourceRuntimeContext:
			value, findErr := runtimeCtx.Find(inputData.SourcePath)
			if findErr != nil {
				return nil, fmt.Errorf("find step input value in initial context: %w", findErr)
			}
			data[inputData.DestinationParam] = value
		default:
			return nil, fmt.Errorf("invalid step input source: %s", inputData.SourceNamespace)
		}
	}
	parsed, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal map: %w", err)
	}
	return &JsonInstanceContext{raw: parsed, data: data}, nil
}

func (c *JsonInstanceContext) GetRaw() json.RawMessage {
	return c.raw
}

func (c *JsonInstanceContext) AppendMap(data map[string]any) (InstanceContext, error) {
	base, ok := c.data.(map[string]any)
	if c.data == nil {
		base = make(map[string]any)
		ok = true
	}
	if !ok {
		return nil, fmt.Errorf("append map: root json value is not an object")
	}

	merged := make(map[string]any, len(base)+len(data))
	for k, v := range base {
		merged[k] = v
	}
	for k, v := range data {
		merged[k] = v
	}

	raw, err := json.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("marshal merged map: %w", err)
	}

	return &JsonInstanceContext{
		raw:  raw,
		data: merged,
	}, nil
}

func (c *JsonInstanceContext) Find(path string) (any, error) {
	if path == "" {
		return nil, fmt.Errorf("path is empty")
	}

	current := c.data
	if current == nil {
		return nil, fmt.Errorf("context is empty")
	}

	parts := strings.Split(path, ".")

	for _, part := range parts {
		switch node := current.(type) {
		case map[string]any:
			value, exists := node[part]
			if !exists {
				return nil, fmt.Errorf("path %q not found: missing key %q", path, part)
			}
			current = value

		case []any:
			idx, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("path %q invalid: %q is not an array index", path, part)
			}
			if idx < 0 || idx >= len(node) {
				return nil, fmt.Errorf("path %q invalid: index %d out of range", path, idx)
			}
			current = node[idx]

		default:
			return nil, fmt.Errorf("path %q not found: %q is not traversable", path, part)
		}
	}

	return current, nil
}
