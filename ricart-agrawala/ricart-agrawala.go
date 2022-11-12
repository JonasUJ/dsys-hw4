// Implements the pseudo-code from slide 15 of lecture 7
// Each function is annotated with its corresponding pseudo-code

package ricartagrawala

type RicartAgrawala interface {
	LamportMut
	SetState(State)
	GetState() State
	Multicast(Request) chan Reply
	Queue() chan Request
	Reply(Request)
}

/*
On initialisation do
	state := RELEASED;
End on
*/
func Init(ra RicartAgrawala) {
	ra.SetState(Released)
}

/*
On enter do
	state := WANTED;
	“multicast ‘req(T,p)’”, where T := LAMPORT time of ‘req’ at p
	wait for N-1 replies
	state := HELD;
End on
*/
func Enter(ra RicartAgrawala, n int) {
	ra.SetState(Wanted)
	replies := ra.Multicast(NewRequest(ra))
	for i := 0; i < n-1; i++ {
		<-replies
	}
	ra.SetState(Held)
}

/*
On receive ‘req (Ti,pi)’do
	if(state == HELD ||
	   (state == WANTED &&
	    (T,pme) < (Ti,pi)))
	then queue req
	else reply to req
End on
*/
func Receive(ra RicartAgrawala, req Request) {
	if ra.GetState() == Held ||
		ra.GetState() == Wanted &&
			Less(ra, req) {
		ra.Queue() <- req
	} else {
		ra.Reply(req)
	}
}

/*
On exit do
	state := RELEASED
	reply to all in queue
End on
*/
func Exit(ra RicartAgrawala) {
	ra.SetState(Released)
    // For some reason this blocks
	// for req := range ra.Queue() {
	//     ra.Reply(req)
	// }
    for {
        select {
        case req := <- ra.Queue():
            ra.Reply(req)
        default:
            return
        }
    }
}
