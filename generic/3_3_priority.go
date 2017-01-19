package generic

import (
	"fmt"

	"golang.org/x/net/http2"

	"github.com/summerwind/h2spec/config"
	"github.com/summerwind/h2spec/spec"
)

func Priority() *spec.TestGroup {
	tg := NewTestGroup("3.3", "PRIORITY")

	// RFC7540, 6.3:
	// The PRIORITY frame (type=0x2) specifies the sender-advised
	// priority of a stream (Section 5.3). It can be sent in any
	// stream state, including idle or closed streams.
	tg.AddTestCase(&spec.TestCase{
		Desc:        "Tests sending a PRIORITY frame with priority 1",
		Requirement: "The endpoint MUST accept PRIORITY frame with priority 1.",
		Run: func(c *config.Config, conn *spec.Conn) error {
			var streamID uint32 = 1

			err := conn.Handshake()
			if err != nil {
				return err
			}

			pp := http2.PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    0,
			}
			conn.WritePriority(streamID, pp)

			headers := spec.CommonHeaders(c)
			hp := http2.HeadersFrameParam{
				StreamID:      streamID,
				EndStream:     true,
				EndHeaders:    true,
				BlockFragment: conn.EncodeHeaders(headers),
			}
			conn.WriteHeaders(hp)

			actual, passed := conn.WaitEventByType(spec.EventHeadersFrame)
			switch event := actual.(type) {
			case spec.HeadersFrameEvent:
				passed = (event.Header().StreamID == streamID)
			default:
				passed = false
			}

			if !passed {
				expected := []string{
					fmt.Sprintf("HEADERS Frame (stream_id:%d)", streamID),
				}

				return &spec.TestError{
					Expected: expected,
					Actual:   actual.String(),
				}
			}

			return nil
		},
	})

	// RFC7540, 6.3:
	// The PRIORITY frame (type=0x2) specifies the sender-advised
	// priority of a stream (Section 5.3). It can be sent in any
	// stream state, including idle or closed streams.
	tg.AddTestCase(&spec.TestCase{
		Desc:        "Tests sending a PRIORITY frame with priority 256",
		Requirement: "The endpoint MUST accept PRIORITY frame with priority 256.",
		Run: func(c *config.Config, conn *spec.Conn) error {
			var streamID uint32 = 1

			err := conn.Handshake()
			if err != nil {
				return err
			}

			pp := http2.PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    255,
			}
			conn.WritePriority(streamID, pp)

			headers := spec.CommonHeaders(c)
			hp := http2.HeadersFrameParam{
				StreamID:      streamID,
				EndStream:     true,
				EndHeaders:    true,
				BlockFragment: conn.EncodeHeaders(headers),
			}
			conn.WriteHeaders(hp)

			actual, passed := conn.WaitEventByType(spec.EventHeadersFrame)
			switch event := actual.(type) {
			case spec.HeadersFrameEvent:
				passed = (event.Header().StreamID == streamID)
			default:
				passed = false
			}

			if !passed {
				expected := []string{
					fmt.Sprintf("HEADERS Frame (stream_id:%d)", streamID),
				}

				return &spec.TestError{
					Expected: expected,
					Actual:   actual.String(),
				}
			}

			return nil
		},
	})

	// RFC7540, 6.3, Stream Dependency:
	// A 31-bit stream identifier for the stream that this stream
	// depends on (see Section 5.3).
	tg.AddTestCase(&spec.TestCase{
		Desc:        "Tests sending a PRIORITY frame with stream dependency",
		Requirement: "The endpoint MUST accept PRIORITY frame with stream dependency.",
		Run: func(c *config.Config, conn *spec.Conn) error {
			var streamID uint32 = 1

			err := conn.Handshake()
			if err != nil {
				return err
			}

			pp := http2.PriorityParam{
				StreamDep: streamID,
				Exclusive: false,
				Weight:    0,
			}
			conn.WritePriority(streamID+2, pp)

			headers := spec.CommonHeaders(c)
			hp := http2.HeadersFrameParam{
				StreamID:      streamID,
				EndStream:     true,
				EndHeaders:    true,
				BlockFragment: conn.EncodeHeaders(headers),
			}
			conn.WriteHeaders(hp)

			actual, passed := conn.WaitEventByType(spec.EventHeadersFrame)
			switch event := actual.(type) {
			case spec.HeadersFrameEvent:
				passed = (event.Header().StreamID == streamID)
			default:
				passed = false
			}

			if !passed {
				expected := []string{
					fmt.Sprintf("HEADERS Frame (stream_id:%d)", streamID),
				}

				return &spec.TestError{
					Expected: expected,
					Actual:   actual.String(),
				}
			}

			return nil
		},
	})

	// RFC7540, 6.3, E:
	// A single-bit flag indicating that the stream dependency is
	// exclusive (see Section 5.3).
	tg.AddTestCase(&spec.TestCase{
		Desc:        "Tests sending a PRIORITY frame with exclusive",
		Requirement: "The endpoint MUST accept PRIORITY frame with exclusive.",
		Run: func(c *config.Config, conn *spec.Conn) error {
			var streamID uint32 = 1

			err := conn.Handshake()
			if err != nil {
				return err
			}

			pp := http2.PriorityParam{
				StreamDep: 0,
				Exclusive: true,
				Weight:    0,
			}
			conn.WritePriority(streamID, pp)

			headers := spec.CommonHeaders(c)
			hp := http2.HeadersFrameParam{
				StreamID:      streamID,
				EndStream:     true,
				EndHeaders:    true,
				BlockFragment: conn.EncodeHeaders(headers),
			}
			conn.WriteHeaders(hp)

			actual, passed := conn.WaitEventByType(spec.EventHeadersFrame)
			switch event := actual.(type) {
			case spec.HeadersFrameEvent:
				passed = (event.Header().StreamID == streamID)
			default:
				passed = false
			}

			if !passed {
				expected := []string{
					fmt.Sprintf("HEADERS Frame (stream_id:%d)", streamID),
				}

				return &spec.TestError{
					Expected: expected,
					Actual:   actual.String(),
				}
			}

			return nil
		},
	})

	return tg
}
