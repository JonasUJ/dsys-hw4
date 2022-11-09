package ricartagrawala

type State uint8

const (
	Released State = iota 
	Wanted
	Held
)

var (
	State_name = map[State]string{
		Released: "Released",
		Wanted: "Wanted",
		Held: "Held",
	}
)

func (state State) String() string {
	return State_name[state]
}
