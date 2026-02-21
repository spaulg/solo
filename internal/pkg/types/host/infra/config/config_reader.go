package config

import (
	domain_config "github.com/spaulg/solo/internal/pkg/impl/host/domain"
)

type ConfigReader interface {
	AddConfigPath(path string) error
	GetConfig() *domain_config.Config
}
