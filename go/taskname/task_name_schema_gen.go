// Code generated by "go run gen_schema.go"; DO NOT EDIT

package taskname

var SCHEMA_FROM_GIT = map[string][]string{
	"Build":       {"os", "compiler", "target_arch", "configuration"},
	"Canary":      {"project", "os", "compiler", "target_arch", "configuration"},
	"Housekeeper": {"frequency"},
	"Infra":       {"frequency"},
	"Perf":        {"os", "compiler", "model", "cpu_or_gpu", "cpu_or_gpu_value", "arch", "configuration"},
	"Test":        {"os", "compiler", "model", "cpu_or_gpu", "cpu_or_gpu_value", "arch", "configuration"},
}

var SEPERATOR_FROM_GIT = "-"
