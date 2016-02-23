//go:generate gogogadget counter.go -p examples -i Counter -c NewCounter -o counter_agent_example.go

package examples

type Counter interface {
	Add(val int64) int64
	Sub(val int64) int64
	Total() int64
}

type counter struct {
	val int64
}

func NewCounter(start int64) Counter {
	return &counter{start}
}

func (c *counter) Add(val int64) int64 {
	c.val += val
	return c.val
}

func (c *counter) Sub(val int64) int64 {
	c.val -= val
	return c.val
}

func (c *counter) Total() int64 {
	return c.val
}
