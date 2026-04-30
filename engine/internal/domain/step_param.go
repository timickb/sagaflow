package domain

import (
	"errors"
	"regexp"
	"strings"
)

var (
	inputDstRegexp    = regexp.MustCompile(`^[a-zA-Z_-][a-zA-Z0-9_-]*$`)
	contextPathRegexp = regexp.MustCompile(`^\$\{(?:runtime|input|result|error)\.[a-zA-Z_-]+(?:\.[a-zA-Z_-]+)*\}$`)
)

type StepParamSource string

const (
	// StepParamSourceInput - значение параметра лежит в initialContext экземпляра
	StepParamSourceInput StepParamSource = "input"
	// StepParamSourceRuntime - значение параметра лежит в runtimeContext экземпляра
	StepParamSourceRuntime StepParamSource = "runtime"
	// StepParamSourceResult - значение параметра лежит в результате выполнения шага
	StepParamSourceResult StepParamSource = "result"
	// StepParamSourceError - значение параметра лежит в информации об ошибке выполнения шага
	StepParamSourceError StepParamSource = "error"
	// StepParamSourceConst - значение параметра задано напрямую
	StepParamSourceConst StepParamSource = "const"
)

type StepInputParam struct {
	// DstVar - с каким названием передать обработчику
	DstVar string
	// StepParamSource - где искать значение
	Source StepParamSource
	// Value - значение
	Value string
}

func NewStepInputParam(dst, src string) (StepInputParam, error) {
	if !inputDstRegexp.MatchString(dst) {
		return StepInputParam{}, errors.New("invalid destination variable")
	}
	if src == "" {
		return StepInputParam{}, errors.New("empty input src")
	}
	matches := contextPathRegexp.FindStringSubmatch(src)
	if matches == nil {
		// не подходит под паттерн пути в контексте -> воспринимаем как константное значение
		return StepInputParam{
			DstVar: dst,
			Source: StepParamSourceConst,
			Value:  src,
		}, nil
	}
	return StepInputParam{
		DstVar: dst,
		Source: StepParamSource(matches[1]),
		Value:  unwrapPathPlaceholder(src),
	}, nil
}

func unwrapPathPlaceholder(s string) string {
	s = strings.TrimPrefix(s, "${")
	s = strings.TrimSuffix(s, "}")
	return s
}
