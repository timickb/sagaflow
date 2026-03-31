package domain

import "encoding/json"

type InstanceContext interface {
	GetRaw() json.RawMessage
}

type JsonInstanceContext struct {
	raw json.RawMessage
}

func NewJsonInstanceContext(raw json.RawMessage) *JsonInstanceContext {
	return &JsonInstanceContext{raw: raw}
}

func (c *JsonInstanceContext) GetRaw() json.RawMessage {
	return c.raw
}
