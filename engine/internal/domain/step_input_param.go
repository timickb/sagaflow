package domain

// StepInputParam - описание параметра, подаваемого на вход шагу
type StepInputParam struct {
	// SourceNamespace - где искать параметр
	SourceNamespace StepInputSource
	// SourcePath - путь к параметру в формате path.to.variable
	SourcePath string
	// DestinationParam - название параметра во входных параметрах шага
	DestinationParam string
}

// StepInputSource - источник параметра для шага
type StepInputSource string

const (
	// StepInputSourceInputContext - параметр из initial контекста саги
	StepInputSourceInputContext StepInputSource = "input"
	// StepInputSourceRuntimeContext - параметр из runtime контекста саги
	StepInputSourceRuntimeContext StepInputSource = "runtime"
)
