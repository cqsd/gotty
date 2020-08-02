package server

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"

	"gopkg.in/segmentio/analytics-go.v3"
)

// wswTap is a tee/tap for a wsWrapper. Technically there's nothing stopping
// you from modifying the stream, but using wsWrapper.getReadTap and
// wsWrapper.getWriteTap to do the tapping will avoid accidentally doing so.
type wswTap func(*wsWrapper) error

// scanLinesCr is bufio.ScanLines but it additionally splits on a lone CR.
//
// Pressing enter on my mac sends a CR, not a LF or CRLF. I'm not sure if this
// is platform-dependent, or if it's an xterm thing, but I think we'd want to
// split on this anyway because a shell will interpret it as a newline.
func scanLinesCr(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// try just splitting like normal first
	advance, token, err = bufio.ScanLines(data, atEOF)

	// no \n or \r\n found, so try finding a \r instead
	if err != nil || advance == 0 {
		if i := bytes.IndexByte(data, '\r'); i >= 0 {
			return i + 1, data[:i], nil
		}
	}

	// i'm not sure if this will ever be true by this point tbh
	if atEOF {
		return len(data), data, nil
	}

	// no new line yet, ask for more data
	return advance, token, err
}

func filterMessageType(reader io.Reader, msgType byte) io.Reader {
	pipeR, pipeW := io.Pipe()
	go func() error {
		buffer := make([]byte, 1024)
		for {
			n, err := reader.Read(buffer)
			if err != nil {
				return err
			}
			if len(buffer) > 0 && buffer[0] == msgType {
				// NB: we chop off the message type here, because we already
				// know what it is at this point. Might not be the right place
				pipeW.Write(buffer[1:n])
			}
		}
	}()
	return pipeR
}

func WithSegment(writeKey string) wswTap {
	return func(wsw *wsWrapper) error {
		client := analytics.New(writeKey)
		log.Printf("Initialized Segment client (%s)\n", writeKey)
		remoteHost := wsw.conn.RemoteAddr().String()
		readTap := wsw.getReadTap()

		go func() {
			// filter for keyboard inputs only. don't emit track calls for
			// terminal resize or pings
			scanner := bufio.NewScanner(filterMessageType(readTap, '1'))
			scanner.Split(scanLinesCr)
			for scanner.Scan() {
				input := scanner.Bytes()
				if len(input) == 0 {
					continue
				}
				s := fmt.Sprintf("%q", input)
				log.Printf("SEGMENT (%s): tracking command %s\n", remoteHost, s)
				client.Enqueue(analytics.Track{
					Event:  "gotty input",
					UserId: remoteHost,
					Properties: analytics.NewProperties().
						Set("input", s),
				})
			}
		}()

		return nil
	}
}

// WithRecordInput tees reads from the websocket (ie data recv'd from the
// client) to a file. *UNIMPLEMENTED*
func WithRecordInput(filename string) wswTap {
	return func(wsw *wsWrapper) error {
		// TODO(cqsd)
		log.Printf("UNIMPLEMENTED: Logging recv'd data to %s\n", filename)
		remoteHost := wsw.conn.RemoteAddr().String()
		inputTap := wsw.getReadTap()
		go func() error {
			buffer := make([]byte, 1024)
			for {
				n, err := inputTap.Read(buffer)
				if err != nil {
					return err
				}
				log.Printf("RECV (%s ): %d bytes (%x)\n", remoteHost, n, buffer[:n])
			}
		}()
		return nil
	}
}

// WithRecordOutput tees writes to the websocket (ie data sent back to the
// client) to a file. *UNIMPLEMENTED*
func WithRecordOutput(filename string) wswTap {
	return func(wsw *wsWrapper) error {
		// TODO(cqsd)
		log.Printf("UNIMPLEMENTED: Logging sent data to %s\n", filename)
		outputTap := wsw.getWriteTap()
		go func() error {
			buffer := make([]byte, 1024)
			for {
				n, err := outputTap.Read(buffer)
				if err != nil {
					return err
				}
				log.Printf("SEND: %d bytes to client (%s)\n", n, string(buffer[:n]))
			}
		}()
		return nil
	}
}
