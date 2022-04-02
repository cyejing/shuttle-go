package channel

type WormholeChannel struct {
}

func NewWormholeChannel(hash []byte) *WormholeChannel {
	return &WormholeChannel{}
}

func (w *WormholeChannel) Run() error {
	return nil
}
