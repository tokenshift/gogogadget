package examples

import . "github.com/tokenshift/gogogadget/lib"

type CounterAgent struct {
	wrapped Counter
	signal chan AgentSignal
	state AgentState

	reqAdd chan struct{val int64}
	resAdd chan struct{val1 int64}

	reqSub chan struct{val int64}
	resSub chan struct{val1 int64}

	reqTotal chan struct{}
	resTotal chan struct{val1 int64}
}

func (agent CounterAgent) Add(val int64) (val1 int64) {
	agent.reqAdd<- struct{val int64}{
		val: val,
	}
	res := <-agent.resAdd
	return res.val1
}

func (agent CounterAgent) Sub(val int64) (val1 int64) {
	agent.reqSub<- struct{val int64}{
		val: val,
	}
	res := <-agent.resSub
	return res.val1
}

func (agent CounterAgent) Total() (val1 int64) {
	agent.reqTotal<- struct{}{
	}
	res := <-agent.resTotal
	return res.val1
}

func (agent CounterAgent) Start() {
	agent.signal<- AGENT_START
}

func (agent CounterAgent) Stop() {
	agent.signal<- AGENT_STOP
}

func (agent CounterAgent) Close() {
	agent.signal<- AGENT_CLOSE
}

func (agent CounterAgent) State() AgentState {
	return agent.state
}

func (agent *CounterAgent) runLoop() {
	for {
		select {
		case signal := <-agent.signal:
			switch signal {
			case AGENT_START:
				agent.state = AGENT_STARTED
			case AGENT_STOP:
				agent.state = AGENT_STOPPED
			case AGENT_CLOSE:
				agent.state = AGENT_CLOSED
				close(agent.reqAdd)
				close(agent.resAdd)
				close(agent.reqSub)
				close(agent.resSub)
				close(agent.reqTotal)
				close(agent.resTotal)
				close(agent.signal)
				return
			}
		case msg := <-agent.reqAdd:
			val1 := agent.wrapped.Add(msg.val)
			agent.resAdd<- struct{val1 int64}{ val1 }
		case msg := <-agent.reqSub:
			val1 := agent.wrapped.Sub(msg.val)
			agent.resSub<- struct{val1 int64}{ val1 }
		case msg := <-agent.reqTotal:
			val1 := agent.wrapped.Total()
			agent.resTotal<- struct{val1 int64}{ val1 }
		}
	}
}

func NewCounterAgent(start int64) CounterAgent {
	wrapped := NewCounter(start)
	agent := CounterAgent{
		wrapped,
		make(chan AgentSignal),
		AGENT_STARTED,
		make(chan struct{val int64}),
		make(chan struct{val1 int64}),
		make(chan struct{val int64}),
		make(chan struct{val1 int64}),
		make(chan struct{}),
		make(chan struct{val1 int64}),
	}

	go agent.runLoop()

	return agent
}

