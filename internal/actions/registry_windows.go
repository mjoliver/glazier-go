//go:build windows

package actions

import (
	"context"
	"fmt"

	"github.com/google/deck"
	glazierReg "github.com/google/glazier/go/registry"
)

func (a *RegistrySet) Run(ctx context.Context) error {
	deck.Infof("registry.set: %s\\%s = %v (type: %s)", a.Config.Path, a.Config.Name, a.Config.Value, a.Config.Type)

	// Ensure key exists
	if err := glazierReg.Create(a.Config.Path); err != nil {
		return fmt.Errorf("registry.set: failed to create key: %w", err)
	}

	switch a.Config.Type {
	case "string":
		return glazierReg.SetString(a.Config.Path, a.Config.Name, fmt.Sprintf("%v", a.Config.Value))
	case "dword":
		var val int
		switch v := a.Config.Value.(type) {
		case int:
			val = v
		case float64:
			val = int(v)
		default:
			return fmt.Errorf("registry.set: dword value must be numeric, got %T", a.Config.Value)
		}
		return glazierReg.SetInteger(a.Config.Path, a.Config.Name, val)
	case "multi_string":
		var vals []string
		if items, ok := a.Config.Value.([]interface{}); ok {
			for _, item := range items {
				vals = append(vals, fmt.Sprintf("%v", item))
			}
		} else {
			return fmt.Errorf("registry.set: multi_string value must be a list")
		}
		return glazierReg.SetMultiString(a.Config.Path, a.Config.Name, vals)
	case "binary":
		if s, ok := a.Config.Value.(string); ok {
			return glazierReg.SetBinary(a.Config.Path, a.Config.Name, []byte(s))
		}
		return fmt.Errorf("registry.set: binary value must be a string")
	default:
		return fmt.Errorf("registry.set: unsupported type: %s", a.Config.Type)
	}
}

func (a *RegistryDelete) Run(ctx context.Context) error {
	deck.Infof("registry.delete: %s\\%s", a.Config.Path, a.Config.Name)
	return glazierReg.Delete(a.Config.Path, a.Config.Name)
}

func (a *RegistryGet) Run(ctx context.Context) error {
	deck.Infof("registry.get: %s\\%s", a.Config.Path, a.Config.Name)

	val, err := glazierReg.GetString(a.Config.Path, a.Config.Name)
	if err != nil {
		return fmt.Errorf("registry.get: %w", err)
	}

	a.Result = val
	deck.Infof("registry.get: %s = %q", a.Config.Name, val)
	return nil
}
