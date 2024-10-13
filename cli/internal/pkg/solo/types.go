package solo

//type StepConfig struct {
//	Name    string `yaml:"name"`
//	Command string `yaml:"command"`
//	Cwd     string `yaml:"cwd"`
//	Timeout int    `yaml:"timeout"`
//}
//
//type StepsConfig struct {
//	Provisioning []StepConfig `yaml:"provisioning"`
//	PreStart     []StepConfig `yaml:"pre_start"`
//	PostStart    []StepConfig `yaml:"post_start"`
//	PreStop      []StepConfig `yaml:"pre_stop"`
//	PostStop     []StepConfig `yaml:"post_stop"`
//	PreDestroy   []StepConfig `yaml:"pre_destroy"`
//	PostDestroy  []StepConfig `yaml:"post_destroy"`
//}
//
//type ServiceConfig struct {
//	Steps StepsConfig `yaml:"steps"`
//}
//
//type Services map[string]ServiceConfig
//
//type ProjectConfig struct {
//	Services Services `yaml:"services"`
//}
