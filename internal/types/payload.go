package types

type Payload interface {
	ToEnvelopePayload() ([]byte, error)
}
