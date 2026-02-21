package config

type OrchestrationConfig struct {
	SearchOrder   []string                      `mapstructure:"search_order" yaml:"search_order"`
	Orchestrators map[string]OrchestratorConfig `mapstructure:"orchestrators" yaml:"orchestrators"`
}
