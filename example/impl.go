package example

type ServiceImpl struct {
}

func (service *ServiceImpl) EchoBytes(msg []byte) ([]byte, error) {
	return msg, nil
}

func (service *ServiceImpl) Echo(msg string) (string, error) {
	return msg, nil
}

func (service *ServiceImpl) Restart() error {
	// Imagine something is being done here
	return nil
}
