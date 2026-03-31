package domain

type SagaHeader struct {
	Name    string
	Version int
}

type SagaView struct {
	SagaHeader
}
