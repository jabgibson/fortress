package fortress

const SharedDataPrefix = "#data:"

type DataOrder struct {
	Order
	Value string `toml:"value"`
}

func (d DataOrder) Sequence() int {
	return d.Seq
}

func (d DataOrder) ExecuteOrder(context OrderContext) (report Report) {
	if d.Value != "" {
		report.Data = map[string]string{
			SharedDataPrefix + d.ID: d.Value,
		}
	}

	return
}

func (d DataOrder) Self() Order {
	return d.Order
}
