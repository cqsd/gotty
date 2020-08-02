package server

import (
	"io"

	"github.com/gorilla/websocket"
)

type wsWrapper struct {
	conn   *websocket.Conn
	reader io.Reader
	writer io.Writer
}

func (wsw *wsWrapper) Write(p []byte) (n int, err error) {
	n, err = wsw.writer.Write(p)
	return n, err
}

func (wsw *wsWrapper) Read(p []byte) (n int, err error) {
	n, err = wsw.reader.Read(p)
	return n, err
}

// getWriteTap "taps" websocket write calls, returning a Reader of all bytes
// that get written to the wsWrapper
func (wsw *wsWrapper) getWriteTap() io.Reader {
	pipeR, pipeW := io.Pipe()
	reader := io.TeeReader(pipeR, wsw.writer)
	wsw.writer = pipeW
	return reader
}

// getReadTap "taps" websocket write calls, returning a Reader of all bytes
// that get written to the wsWrapper
func (wsw *wsWrapper) getReadTap() io.Reader {
	pipeR, pipeW := io.Pipe()
	reader := io.TeeReader(wsw.reader, pipeW)
	wsw.reader = pipeR
	return reader
}

// make wswrapper have its reader and writer be their own fields, allowing us
// to alter the wswrapper's read and write behavior at runtime by replacing
// them with custom readers/writers
type wsWriter struct {
	*websocket.Conn
}

func (wsw *wsWriter) Write(p []byte) (n int, err error) {
	writer, err := wsw.Conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return 0, err
	}
	defer writer.Close()
	return writer.Write(p)
}

type wsReader struct {
	*websocket.Conn
}

func (wsr *wsReader) Read(p []byte) (n int, err error) {
	for {
		msgType, reader, err := wsr.Conn.NextReader()
		if err != nil {
			return 0, err
		}

		if msgType != websocket.TextMessage {
			continue
		}

		return reader.Read(p)
	}
}

func NewWsWrapper(conn *websocket.Conn) *wsWrapper {
	return &wsWrapper{
		conn:   conn,
		reader: &wsReader{conn},
		writer: &wsWriter{conn},
	}
}
