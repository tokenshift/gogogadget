//go:generate gogogadget agent Counter -c NewCounter -i counter.go -I -p examples

package examples

type Counter interface {
	Add(int64) int64
	Sub(int64) int64
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