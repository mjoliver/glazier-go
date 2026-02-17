package actions

import (
	"context"
	"fmt"
	"log"

	join "github.com/google/glazier/go/domain"
	"gopkg.in/yaml.v3"
)

type DomainJoinConfig struct {
	Domain   string `yaml:"domain"`
	OU       string `yaml:"ou"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

func NewDomainJoin(ctx context.Context, yamlData interface{}) (Action, error) {
	var cfg DomainJoinConfig
	data, err := yaml.Marshal(yamlData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &DomainJoin{Config: cfg}, nil
}

type DomainJoin struct {
	Config DomainJoinConfig
}

func (a *DomainJoin) Run(ctx context.Context) error {
	log.Printf("Joining domain: %s (OU: %s)", a.Config.Domain, a.Config.OU)

	// Call library
	return join.Domain(a.Config.Domain, a.Config.OU, a.Config.User, a.Config.Password, join.JoinDomainFlag)
}

func (a *DomainJoin) Validate() error {
	if a.Config.Domain == "" {
		return fmt.Errorf("domain is required")
	}
	return nil
}

func init() {
	Register("domain.join", NewDomainJoin)
}
