package ricartagrawala

type Reply struct {}

type Request struct {
    Time uint64
    Pid uint32
}

// Create a new request with the time and pid of the passed lamport
func NewRequest(lamport Lamport) Request {
	return Request{
		Time:    lamport.GetTime(),
		Pid:     lamport.GetPid(),
	}
}

func (req Request) GetTime() uint64 {
	return req.Time
}

func (req Request) GetPid() uint32 {
	return req.Pid
}

func (req Request) SetTime(time uint64) {
	req.Time = time
}
