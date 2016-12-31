package http2

import (
	"fmt"
	"reflect"

	"golang.org/x/net/http2"

	"github.com/summerwind/h2spec/config"
	"github.com/summerwind/h2spec/spec"
)

func Ping() *spec.TestGroup {
	tg := NewTestGroup("6.7", "PING")

	// Receivers of a PING frame that does not include an ACK flag MUST
	// send a PING frame with the ACK flag set in response, with an
	// identical payload.
	tg.AddTestCase(&spec.TestCase{
		Desc:        "Sends a PING frame",
		Requirement: "The endpoint MUST sends a PING frame with ACK, with an identical payload.",
		Run: func(c *config.Config, conn *spec.Conn) error {
			err := conn.Handshake()
			if err != nil {
				return err
			}

			data := [8]byte{'h', '2', 's', 'p', 'e', 'c'}
			conn.WritePing(false, data)

			actual, passed := conn.WaitEventByType(spec.EventPingFrame)
			switch event := actual.(type) {
			case spec.PingFrameEvent:
				if event.IsAck() && reflect.DeepEqual(event.Data, data) {
					passed = true
				}
			default:
				passed = false
			}

			if !passed {
				expected := []string{
					"PING Frame (length:8, flags:0x01, stream_id:0)",
				}

				return &spec.TestError{
					Expected: expected,
					Actual:   actual.String(),
				}
			}

			return nil
		},
	})

	// ACK (0x1):
	// When set, bit 0 indicates that this PING frame is a PING
	// response. An endpoint MUST set this flag in PING responses.
	// An endpoint MUST NOT respond to PING frames containing this
	// flag.
	tg.AddTestCase(&spec.TestCase{
		Desc:        "Sends a PING frame with ACK",
		Requirement: "The endpoint MUST NOT respond to PING frames with ACK.",
		Run: func(c *config.Config, conn *spec.Conn) error {
			err := conn.Handshake()
			if err != nil {
				return err
			}

			unexpectedData := [8]byte{'i', 'n', 'v', 'a', 'l', 'i', 'd'}
			expectedData := [8]byte{'h', '2', 's', 'p', 'e', 'c'}
			conn.WritePing(true, unexpectedData)
			conn.WritePing(false, expectedData)

			actual, passed := conn.WaitEventByType(spec.EventPingFrame)
			switch event := actual.(type) {
			case spec.PingFrameEvent:
				if reflect.DeepEqual(event.Data, unexpectedData) {
					passed = false
				} else if event.IsAck() && reflect.DeepEqual(event.Data, expectedData) {
					passed = true
				}
			default:
				passed = false
			}

			if !passed {
				var actualStr string

				expected := []string{
					fmt.Sprintf("PING Frame (opaque_data: %s)", expectedData),
				}

				f, ok := actual.(spec.PingFrameEvent)
				if ok {
					actualStr = fmt.Sprintf("PING Frame (opaque_data: %s)", f.Data)
				} else {
					actualStr = actual.String()
				}

				return &spec.TestError{
					Expected: expected,
					Actual:   actualStr,
				}
			}

			return nil
		},
	})

	// If a PING frame is received with a stream identifier field value
	// other than 0x0, the recipient MUST respond with a connection
	// error (Section 5.4.1) of type PROTOCOL_ERROR.
	tg.AddTestCase(&spec.TestCase{
		Desc:        "Sends a PING frame with a stream identifier field value other than 0x0",
		Requirement: "The endpoint MUST respond with a connection error of type PROTOCOL_ERROR.",
		Run: func(c *config.Config, conn *spec.Conn) error {
			err := conn.Handshake()
			if err != nil {
				return err
			}

			// PING frame:
			// length: 8, flags: 0x0, stream_id: 1
			conn.Send([]byte("\x00\x00\x08\x06\x00\x00\x00\x00\x01"))
			conn.Send([]byte("\x00\x00\x00\x00\x00\x00\x00\x00"))

			return spec.VerifyConnectionError(conn, http2.ErrCodeProtocol)
		},
	})

	// Receipt of a PING frame with a length field value other than 8
	// MUST be treated as a connection error (Section 5.4.1) of type
	// FRAME_SIZE_ERROR.
	tg.AddTestCase(&spec.TestCase{
		Desc:        "Sends a PING frame with a length field value other than 8",
		Requirement: "The endpoint MUST treat this as a connection error of type FRAME_SIZE_ERROR.",
		Run: func(c *config.Config, conn *spec.Conn) error {
			err := conn.Handshake()
			if err != nil {
				return err
			}

			// PING frame:
			// length: 8, flags: 0x0, stream_id: 1
			conn.Send([]byte("\x00\x00\x06\x06\x00\x00\x00\x00\x01"))
			conn.Send([]byte("\x00\x00\x00\x00\x00\x00"))

			return spec.VerifyConnectionError(conn, http2.ErrCodeFrameSize)
		},
	})

	return tg
}
