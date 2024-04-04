package utils

import (
	bytesize "github.com/inhies/go-bytesize"
)

func computeSenders(toSend bytesize.ByteSize, minSendSize bytesize.ByteSize) []bytesize.ByteSize {
	numSenders := uint64(toSend / minSendSize)
	senders := make([]bytesize.ByteSize, numSenders)
	var i uint64
	for i = 0; i < numSenders; i++ {
		senders[i] = bytesize.ByteSize(minSendSize)
	}
	leftover := bytesize.ByteSize(uint64(toSend) - (numSenders * uint64(minSendSize)))
	if leftover != 0 {
		senders = append(senders, leftover)
	}
	return senders
}

// ComputeSenderSizes returns a slice of bytesize.ByteSize, where each elment represents a
// client/sender, and the value represents how much data that client/sender should send, such
// that a total of toSend can be sent without exceeding maxSenders.
func ComputeSenderSizes(maxSenders int, toSend bytesize.ByteSize, minSendSize bytesize.ByteSize) []bytesize.ByteSize {
	var senders = computeSenders(toSend, minSendSize)
	inc := minSendSize
	for len(senders) > maxSenders {
		minSendSize += inc
		senders = computeSenders(toSend, minSendSize)
	}
	return senders
}
