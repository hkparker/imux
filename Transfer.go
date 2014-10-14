//package multiplexity
package main

// transfer interface for FileTransfer and DirectoryTransfer structs?

type Transfer struct {
	Source Host
	Destination Host
	SourcePath string
	SourceName string
	DestinationPath string
	DestinationName string
	SizeBytes int
	TransferredBytes int
}
