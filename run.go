package fortress

type RunOrder struct {
	Order
	Command string   `toml:"command"`
	Args    []string `toml:"args"`
	Find    bool     `toml:"find"`
}

func (r RunOrder) ExecuteOrder(context OrderContext) Report {

	return Report{} //TODO implement
}

func (r RunOrder) Sequence() int {
	return r.Seq
}

func (r RunOrder) Self() Order {
	return r.Order
}
