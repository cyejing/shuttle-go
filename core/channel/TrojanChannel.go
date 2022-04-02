package channel

type TrojanChannel struct {

}

func NewTrojanChannel() *TrojanChannel {
	return &TrojanChannel{}
}

func (t *TrojanChannel) Run() error  {
	return nil
}
