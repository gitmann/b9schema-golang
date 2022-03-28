package threeflag

import "fmt"

// ThreeFlag implements a 3-value flag: "undefined", "true", "false"
type ThreeFlag int

const (
	Undefined ThreeFlag = 0
	False     ThreeFlag = -1
	True      ThreeFlag = 1
)

func (tf ThreeFlag) String() string {
	return fmt.Sprintf("%d", tf)
}
