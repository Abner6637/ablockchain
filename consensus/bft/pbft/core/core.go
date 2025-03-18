package pbftcore

type PBFT struct {
	p2pNodeID string
	view      uint64
	round     uint64
}

func (p *PBFT) Start() error {
	return nil
}

func (p *PBFT) Stop() error {
	return nil
}
