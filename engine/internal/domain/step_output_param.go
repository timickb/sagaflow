package domain

// StepOutputParam - описание параметра, ожидаемого в результате выполнения шага.
// Всегда сохраняется в runtime контекст саги.
type StepOutputParam struct {
	// SourceNamespace - в каком разделе результата шага искать параметр
	SourceNamespace StepOutputSource
	// SourceParam - название параметра из результата шага
	SourceParam string
	// DestinationParam - название параметра для сохранения в runtime контекст саги
	DestinationParam string
}

// StepOutputSource - где в результате шага искать параметр
type StepOutputSource string

const (
	// StepOutputSourceResult - секция result из события результата шага
	StepOutputSourceResult StepOutputSource = "result"
	// StepOutputSourceError - секция error из события результата шага
	StepOutputSourceError StepOutputSource = "error"
)
