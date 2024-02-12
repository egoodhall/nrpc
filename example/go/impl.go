package example

import "log/slog"

type ServiceImpl struct {
}

func (service *ServiceImpl) Echo(msg *EchoRequest) (*EchoReply, error) {
	slog.Info("Handling request", "message", msg.Message)
	return &EchoReply{Message: msg.Message}, nil
}
