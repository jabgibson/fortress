package fortress

type Order struct {
	Seq          int    `toml:"_seq"`
	ID           string `toml:"id"`
	IgnoreFail   bool   `toml:"ignore-fail"`
	IgnoreGlobal bool   `toml:"ignore-global"`
}

type Orderer interface {
	ExecuteOrder(context OrderContext) Report
	Sequence() int
	Self() Order
}

type Report struct {
	EnvDirections []EnvDirection
	Data          map[string]string
	ExitCode      int
	Errors        []error
	Output        []byte
}

type OrderContext struct {
	Owner   string
	EnvVars map[string]string
	Data    map[string]string
}

// Sequence sorting
type BySequence []Orderer

func (s BySequence) Len() int {
	return len(s)
}
func (s BySequence) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s BySequence) Less(i, j int) bool {
	return s[i].Sequence() < s[j].Sequence()
}
