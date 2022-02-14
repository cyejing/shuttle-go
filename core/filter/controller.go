package filter

type controller struct {

}

func (s controller) Init() {
}

func (s controller) Name() string {
	return "controller"
}

func (s controller) Filter(exchange *Exchange, c interface{}) error {
	panic("implement me")
}



