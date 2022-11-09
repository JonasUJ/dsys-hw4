package ricartagrawala

type Lamport interface {
	GetTime() uint64
	GetPid() uint32
}

type LamportMut interface {
	Lamport
	SetTime(uint64)
}

// Calculate new time on a send event.
func Send(lamport LamportMut) {
	lamport.SetTime(lamport.GetTime() + 1)
}

// Calculate new time on a recv event, comparing the two lamports to determine which is greater.
func Recv(lamport LamportMut, other Lamport) {
	if Less(lamport, other) {
		lamport.SetTime(other.GetTime() + 1)
	} else {
		lamport.SetTime(lamport.GetTime() + 1)
	}
}

// Compare two lamports (according to the spec ofc. ;)
func Less(lamport, other Lamport) bool {
	// First compare by time, then by pid (lower pid is greater)
	return lamport.GetTime() < other.GetTime() ||
		lamport.GetTime() == other.GetTime() &&
			lamport.GetPid() > other.GetPid()
}
