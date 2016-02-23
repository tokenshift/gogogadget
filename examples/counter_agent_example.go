package examples

import . "github.com/tokenshift/gogogadget/lib"

type CounterAgent struct {
	wrapped Counter
	signal chan AgentSignal
	state AgentState

	reqAdd chan struct{int64}
	resAdd chan struct{int64}

	reqSub chan struct{int64}
	resSub chan struct{int64}

	reqTotal chan struct{}
	resTotal chan struct{int64}
}

func NewCounterAgent(start int64) CounterAgent {
	agent := CounterAgent {
		NewCounter(start),
		make(chan AgentSignal),
		AGENT_STARTED,
		make(chan struct{int64}),
		make(chan struct{int64}),
		make(chan struct{int64}),
		make(chan struct{int64}),
		make(chan struct{}),
		make(chan struct{int64}),
	}

	go agent.runLoop()

	return agent
}

func (c CounterAgent) Add(val int64) int64 {
	c.reqAdd <- struct{int64}{val}
	res := <- c.resAdd
	return res.int64
}

func (c CounterAgent) Sub(val int64) int64 {
	c.reqSub <- struct{int64}{val}
	res := <- c.resSub
	return res.int64
}

func (c CounterAgent) Total() int64 {
	c.reqTotal <- struct{}{}
	res := <- c.resTotal
	return res.int64
}

func (c *CounterAgent) runLoop() {
	for {
		select {
		case signal := <-c.signal:
			switch signal {
			case AGENT_START:
				c.state = AGENT_STARTED
			case AGENT_STOP:
				c.state = AGENT_STOPPED
			case AGENT_CLOSE:
				c.state = AGENT_CLOSED
				close(c.reqAdd)
				close(c.resAdd)
				close(c.reqSub)
				close(c.resSub)
				close(c.reqTotal)
				close(c.resTotal)
				close(c.signal)
				return
			}
		case msg := <-c.reqAdd:
			c.resAdd<- struct{int64}{c.wrapped.Add(msg.int64)}
		case msg := <-c.reqSub:
			c.resSub<- struct{int64}{c.wrapped.Sub(msg.int64)}
		case _ = <-c.reqTotal:
			c.resTotal<- struct{int64}{c.wrapped.Total()}
		}
	}
}

func (c CounterAgent) Start() {
	c.signal <- AGENT_START
}

func (c CounterAgent) Stop() {
	c.signal <- AGENT_STOP
}

func (c CounterAgent) Close() {
	c.signal <- AGENT_CLOSE
}

func (c CounterAgent) State() AgentState {
	return c.state
}
