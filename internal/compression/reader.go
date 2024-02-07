package compression

import "io"

type peekReader struct {
	src    io.Reader
	buffer []byte
	offset int
}

func newPeekReader(src io.Reader) PeekReader {
	return &peekReader{
		src:    src,
		buffer: make([]byte, 0, 6),
	}
}

// Peek returns the first n bytes without advancing the reader.
// It reads more data into the buffer if necessary.
func (p *peekReader) Peek(n int) ([]byte, error) {
	// Ensure the buffer has enough data
	if len(p.buffer)-p.offset < n {
		// Read more data: calculate how much more is needed
		needed := n - (len(p.buffer) - p.offset)
		tempBuf := make([]byte, needed)
		read, err := io.ReadFull(p.src, tempBuf)
		if err != nil && err != io.EOF {
			return nil, err // Propagate read errors (except EOF)
		}
		// Append any new data to the buffer
		p.buffer = append(p.buffer, tempBuf[:read]...)
	}

	// If after reading there's still not enough data for peek, but no more data is coming
	if len(p.buffer)-p.offset < n {
		return nil, io.EOF // Or another error indicating not enough data
	}

	// Return the requested data without advancing the offset
	return p.buffer[p.offset : p.offset+n], nil
}

// Read reads data into p, advancing the reader.
func (p *peekReader) Read(b []byte) (int, error) {
	// If there's buffered data, use it
	if p.offset < len(p.buffer) {
		n := copy(b, p.buffer[p.offset:])
		p.offset += n
		// If the buffer is fully consumed, reset it
		if p.offset == len(p.buffer) {
			p.buffer = p.buffer[:0]
			p.offset = 0
		}
		return n, nil
	}

	// No buffered data; read directly from the source
	return p.src.Read(b)
}
