package domain

import (
	"encoding/json"
	"fmt"
)

type InstanceContext interface {
	GetRaw() json.RawMessage
	AppendMap(data map[string]any) (InstanceContext, error)
}

type JsonInstanceContext struct {
	raw json.RawMessage
}

func NewJsonInstanceContextFromRaw(raw json.RawMessage) *JsonInstanceContext {
	return &JsonInstanceContext{raw: raw}
}

func NewJsonInstanceContextFromAny(data any) (*JsonInstanceContext, error) {
	parsed, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal map: %w", err)
	}
	return &JsonInstanceContext{raw: parsed}, nil
}

func (c *JsonInstanceContext) GetRaw() json.RawMessage {
	return c.raw
}

func (c *JsonInstanceContext) AppendMap(data map[string]any) (InstanceContext, error) {
	base := make(map[string]any)

	if len(c.raw) > 0 && string(c.raw) != "null" {
		if err := json.Unmarshal(c.raw, &base); err != nil {
			return nil, fmt.Errorf("unmarshal raw message: %w", err)
		}
	}

	for k, v := range data {
		base[k] = v
	}

	merged, err := json.Marshal(base)
	if err != nil {
		return nil, fmt.Errorf("marshal merged map: %w", err)
	}

	return NewJsonInstanceContextFromRaw(merged), nil
}
