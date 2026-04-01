package dsl

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/engine/pkg/utils"
	"gopkg.in/yaml.v3"
)

var (
	allowedDSLExtensions = []string{"yaml", "yml"}
)

type Cache struct {
	definitions map[domain.SagaDefinitionHeader]*domain.SagaDefinition
}

func NewCache(dirPath string) (*Cache, error) {
	rawSagas, err := readRawSagas(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read definitions: %w", err)
	}
	definitions := make(map[domain.SagaDefinitionHeader]*domain.SagaDefinition, len(rawSagas))
	for _, rawSaga := range rawSagas {
		definition, vErr := rawSaga.ValidateAndNormalize()
		if vErr != nil {
			return nil, fmt.Errorf("failed to validate saga definition: %w", vErr)
		}
		header := domain.SagaDefinitionHeader{
			Name:    definition.Name,
			Version: definition.Version,
		}
		definitions[header] = definition
	}
	return &Cache{definitions: definitions}, nil
}

func (r *Cache) GetSagaDefinition(header domain.SagaDefinitionHeader) (*domain.SagaDefinition, bool) {
	view, ok := r.definitions[header]
	return view, ok
}

func readRawSagas(path string) ([]*RawSagaDefinition, error) {
	entries, _ := os.ReadDir(path)
	handledSagas := make([]*RawSagaDefinition, 0)

	for _, entry := range entries {
		if entry.IsDir() {
			children, err := readRawSagas(path + "/" + entry.Name())
			if err != nil {
				return nil, err
			}
			handledSagas = append(handledSagas, children...)
		} else {
			scenario, err := readRawSaga(path, entry)
			if err != nil {
				return nil, err
			}
			handledSagas = append(handledSagas, scenario)
		}
	}

	return handledSagas, nil
}

func readRawSaga(path string, entry os.DirEntry) (*RawSagaDefinition, error) {
	nameParts := strings.Split(entry.Name(), ".")
	if len(nameParts) < 3 {
		return nil, errors.New("invalid DSL file name: correct template is <saga_name>.<saga_version>.yaml|yml")
	}
	if !utils.Contains(allowedDSLExtensions, nameParts[len(nameParts)-1]) {
		return nil, fmt.Errorf("invalid DSL file extension: %s", nameParts[len(nameParts)-1])
	}

	content, err := os.ReadFile(path + "/" + entry.Name())
	if err != nil {
		return nil, err
	}

	var raw RawSagaDefinition
	if err = yaml.Unmarshal(content, &raw); err != nil {
		return nil, err
	}
	return &raw, nil
}
