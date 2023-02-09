package dendrite

import (
	"context"
	"dendrite/internal/pkg/backend"
	"dendrite/internal/pkg/dendrite/dto"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/tmc/graphql"
	"github.com/tmc/graphql/parser"
)

type DendriteService struct {
	backend backend.Backend
}

func NewDendriteService(backend backend.Backend) *DendriteService {
	return &DendriteService{
		backend: backend,
	}
}

// -1: version not found --> current
func (s *DendriteService) GetFieldVersion(args []graphql.Argument) (int, error) {
	for _, arg := range args {
		if arg.Name == "version" {
			switch arg.Value.(type) {
			case int:
				return arg.Value.(int), nil
			default:
				return 0, errors.New("invalid version provided")
			}
		}
	}
	return -1, nil
}

func (s *DendriteService) GetSelectionsByField(field graphql.Field, base string) ([]dto.Selection, error) {
	if len(field.SelectionSet) <= 0 {
		version, err := s.GetFieldVersion(field.Arguments)
		if err != nil {
			return nil, err
		}
		return []dto.Selection{{
			Path:    path.Join(base, field.Name),
			Version: version,
		}}, nil
	} else {
		output := []dto.Selection{}
		for _, section := range field.SelectionSet {
			urls, err := s.GetSelectionsByField(*section.Field, fmt.Sprintf("%v/%v", base, field.Name))
			if err != nil {
				return nil, err
			}
			output = append(output, urls...)
		}
		return output, nil
	}
}

func (s *DendriteService) GetConfigsBySelection(ctx context.Context, selection dto.Selection) ([]string, error) {
	if selection.Version == -1 {
		result, err := s.backend.GetManyCurrent(ctx, selection.Path)
		if err != nil {
			return nil, err
		}
		return result, nil
	} else {
		result, err := s.backend.GetMany(ctx, selection.Path, selection.Version)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
}

func (s *DendriteService) GetObjectByPaths(ctx context.Context, selections []dto.Selection) (map[string]any, error) {
	output := make(map[string]any)
	for _, selection := range selections {
		if !strings.HasPrefix(selection.Path, "/") {
			return nil, errors.New("invalid path")
		}
		values, err := s.GetConfigsBySelection(ctx, selection)
		if err != nil {
			return nil, fmt.Errorf("failed to get values from db: %w", err)
		}
		if len(values) > 0 {
			var u = output
			subPaths := strings.Split(selection.Path, "/")[1:]
			for i, subPath := range subPaths {
				if i == len(subPaths)-1 {
					switch u[subPath].(type) {
					case string:
						u[subPath] = append(values, u[subPath].(string))
					case []string:
						u[subPath] = append(u[subPath].([]string), values...)
					case map[string]any:
						if len(values) > 1 {
							u[subPath].(map[string]any)["/"] = values
						} else {
							u[subPath].(map[string]any)["/"] = values[0]
						}
					case nil:
						if len(values) > 1 {
							u[subPath] = values
						} else {
							u[subPath] = values[0]
						}
					}
				} else {
					if u[subPath] == nil {
						u[subPath] = make(map[string]any)
					}
					u = u[subPath].(map[string]any)
				}
			}
		}

	}
	return output, nil
}

func (s *DendriteService) Query(ctx context.Context, query string) (map[string]any, error) {
	operation, err := parser.ParseOperation([]byte(query))
	if err != nil {
		return nil, err
	}
	selections := []dto.Selection{}
	for _, section := range operation.SelectionSet {
		output, err := s.GetSelectionsByField(*section.Field, "/")
		if err != nil {
			return nil, err
		}
		selections = append(selections, output...)
	}
	object, err := s.GetObjectByPaths(ctx, selections)
	if err != nil {
		return nil, err
	}
	return object, nil
}
