package example

type ServiceImpl struct {
}

func (service *ServiceImpl) Echo(msg *EchoRequest) (*EchoReply, error) {
	return &EchoReply{Message: msg.Message}, nil
}
