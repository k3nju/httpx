package httpx

type Message interface {
	HeaderBytes() []byte
	BodyReader() BodyReader
}
