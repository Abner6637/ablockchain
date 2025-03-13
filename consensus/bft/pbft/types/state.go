package pbfttypes

type State uint64

const (
	StateAcceptRequest State = iota
	StatePreprepared
	StatePrepared
	StateCommitted
)

// 比较两个状态的先后顺序
func (s State) Cmp(y State) int {
	if uint64(s) < uint64(y) {
		return -1
	}
	if uint64(s) > uint64(y) {
		return 1
	}
	return 0
}
