package rafttypes

type State int

const (
	StateFollower State = iota
	StateCandidate
	StateLeader
)
