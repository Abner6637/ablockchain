package pbfttypes

const (
	MsgPreprepare uint64 = iota
	MsgPrepare
	MsgCommit
	MsgRoundChange
	// msgAll
)

type Message struct {
	Code      uint64
	Msg       []byte
	Address   []byte
	Signature []byte
}
