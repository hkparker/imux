package imux

type Chunk struct {
	SocketID   string
	SequenceID int
	Data       []byte
}
