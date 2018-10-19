package fortress

import (
	"os"
)

type EnvOrder struct {
	Order
	Key     string   `toml:"key"`
	Value   string   `toml:"value"`
	Targets []string `toml:"targets"`
}

func (e EnvOrder) ExecuteOrder(context OrderContext) (report Report) {
	var envDirections []EnvDirection

	// If no targets are defined, variable is global scope
	if len(e.Targets) == 0 {
		os.Setenv(e.Key, e.Value)
		envDirections = append(envDirections, EnvDirection{
			Target: "#GLOBAL#",
			Key:    e.Key,
			Value:  e.Value,
		})
	} else {
		for _, target := range e.Targets {
			envDirections = append(envDirections, EnvDirection{
				Target: target,
				Key:    e.Key,
				Value:  e.Value,
			})
		}
	}
	report.EnvDirections = envDirections
	return
}

func (e EnvOrder) Sequence() int {
	return e.Seq
}

func (e EnvOrder) Self() Order {
	return e.Order
}

type EnvDirection struct {
	Target string
	Key    string
	Value  string
}
