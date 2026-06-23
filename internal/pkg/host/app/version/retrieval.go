package version

import (
	domain_version "github.com/spaulg/solo/internal/pkg/common/domain/version"
)

func Get() domain_version.Info {
	return domain_version.Get()
}
