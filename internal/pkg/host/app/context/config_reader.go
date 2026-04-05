package context

import (
	domain_config "github.com/spaulg/solo/internal/pkg/host/domain"
)

type ConfigReader interface {
	AddConfigPath(path string) error
	GetConfig() *domain_config.Config
}
