package main

type Chunk struct {
	Data []byte
	ID int
}

func (chunk *Chunk) GenerateHeader() (header []byte, err error) {
	var buffer bytes.Buffer
	chunk_id := strconv.Itoa(chunk.ID)
	chunk_size := strconv.Itoa(len(chunk.Data))
	buffer.WriteString(chunk_id)
	buffer.WriteString(" ")
	buffer.WriteString(chunk_size)
	space := 32-len(header)
	for i := 0; i < space; i++ {
		buffer.WriteString(" ")
	}
	header := buffer.Bytes()
	if len(header) > 32 {
		err := errors.New("Chunk header length exceeds 32 byte limit")
	}
	return
}
