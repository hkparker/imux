package imux

var ClientChunkSize = 16384
var ObservedMaximums = make(map[string]int)

func GetChunkSize(chunk_size_mode string) int {
	if chunk_size_mode == "client" {
		return ClientChunkSize
	} else if chunk_size_mode == "server" {

	}
	return 16384
}
