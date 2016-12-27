package http2

import (
	"fmt"

	"golang.org/x/net/http2"

	"github.com/summerwind/h2spec/config"
	"github.com/summerwind/h2spec/spec"
)

func InitialFlowControlWindowSize() *spec.TestGroup {
	tg := NewTestGroup("6.9.2", "Initial Flow-Control Window Size")

	// When the value of SETTINGS_INITIAL_WINDOW_SIZE changes,
	// a receiver MUST adjust the size of all stream flow-control
	// windows that it maintains by the difference between the new
	// value and the old value.
	tg.AddTestCase(&spec.TestCase{
		Desc:        "Changes SETTINGS_INITIAL_WINDOW_SIZE after sending HEADERS frame",
		Requirement: "The endpoint MUST adjust the size of all stream flow-control windows.",
		Run: func(c *config.Config, conn *spec.Conn) error {
			var streamID uint32 = 1
			var actual spec.Event

			err := conn.Handshake()
			if err != nil {
				return err
			}

			headers := spec.CommonHeaders(c)
			hp1 := http2.HeadersFrameParam{
				StreamID:      streamID,
				EndStream:     true,
				EndHeaders:    true,
				BlockFragment: conn.EncodeHeaders(headers),
			}
			conn.WriteHeaders(hp1)

			// Get the length of response body.
			resLen := -1
			for resLen == -1 {
				ev := conn.WaitEvent()

				switch event := ev.(type) {
				case spec.EventDataFrame:
					resLen = int(event.Header().Length)
				}
			}

			// Skip this test case when the length of response body is 0.
			if resLen < 1 {
				return spec.ErrSkipped
			}

			// Set SETTINGS_INITIAL_WINDOW_SIZE to 0 to prevent sending
			// DATA frame.
			settings1 := []http2.Setting{
				http2.Setting{
					ID:  http2.SettingInitialWindowSize,
					Val: 0,
				},
			}
			conn.WriteSettings(settings1...)

			err = spec.VerifyFrameType(conn, http2.FrameSettings)
			if err != nil {
				return err
			}

			// Send a HEADERS frame.
			streamID += 2
			hp2 := http2.HeadersFrameParam{
				StreamID:      streamID,
				EndStream:     true,
				EndHeaders:    true,
				BlockFragment: conn.EncodeHeaders(headers),
			}
			conn.WriteHeaders(hp2)

			// Set SETTINGS_INITIAL_WINDOW_SIZE to 1 so that the server
			// can send DATA frame.
			settings2 := []http2.Setting{
				http2.Setting{
					ID:  http2.SettingInitialWindowSize,
					Val: 1,
				},
			}
			conn.WriteSettings(settings2...)

			err = spec.VerifyFrameType(conn, http2.FrameSettings)
			if err != nil {
				return err
			}

			// Wait for DATA frame...
			passed := false
			for !conn.Closed {
				ev := conn.WaitEvent()

				switch event := ev.(type) {
				case spec.EventDataFrame:
					actual = event
					passed = (event.Header().Length == 1)
				case spec.EventTimeout:
					if actual == nil {
						actual = event
					}
				default:
					actual = ev
				}

				if passed {
					break
				}
			}

			if !passed {
				expected := []string{
					fmt.Sprintf("DATA Frame (length:1, flags:0x00, stream_id:%d)", streamID),
				}

				return &spec.TestError{
					Expected: expected,
					Actual:   actual.String(),
				}
			}

			return nil
		},
	})

	// A sender MUST track the negative flow-control window and
	// MUST NOT send new flow-controlled frames until it receives
	// WINDOW_UPDATE frames that cause the flow-control window to
	// become positive.
	tg.AddTestCase(&spec.TestCase{
		Desc:        "Sends a SETTINGS frame for window size to be negative",
		Requirement: "The endpoint MUST track the negative flow-control window.",
		Run: func(c *config.Config, conn *spec.Conn) error {
			var streamID uint32 = 1
			var actual spec.Event

			err := conn.Handshake()
			if err != nil {
				return err
			}

			headers := spec.CommonHeaders(c)
			hp1 := http2.HeadersFrameParam{
				StreamID:      streamID,
				EndStream:     true,
				EndHeaders:    true,
				BlockFragment: conn.EncodeHeaders(headers),
			}
			conn.WriteHeaders(hp1)

			// Get the length of response body.
			resLen := -1
			for resLen == -1 {
				ev := conn.WaitEvent()

				switch event := ev.(type) {
				case spec.EventDataFrame:
					resLen = int(event.Header().Length)
				}
			}

			// Skip this test case when the length of response body is 0.
			if resLen < 5 {
				return spec.ErrSkipped
			}

			// Set SETTINGS_INITIAL_WINDOW_SIZE to 3 to prevent sending
			// all of DATA frame.
			settings1 := []http2.Setting{
				http2.Setting{
					ID:  http2.SettingInitialWindowSize,
					Val: 3,
				},
			}
			conn.WriteSettings(settings1...)

			err = spec.VerifyFrameType(conn, http2.FrameSettings)
			if err != nil {
				return err
			}

			// Send a HEADERS frame.
			streamID += 2
			hp2 := http2.HeadersFrameParam{
				StreamID:      streamID,
				EndStream:     true,
				EndHeaders:    true,
				BlockFragment: conn.EncodeHeaders(headers),
			}
			conn.WriteHeaders(hp2)

			// Verify reception of DATA frame.
			err = spec.VerifyFrameType(conn, http2.FrameData)
			if err != nil {
				return err
			}

			// Set SETTINGS_INITIAL_WINDOW_SIZE to 2 to make the window
			// size negative.
			settings2 := []http2.Setting{
				http2.Setting{
					ID:  http2.SettingInitialWindowSize,
					Val: 2,
				},
			}
			conn.WriteSettings(settings2...)

			err = spec.VerifyFrameType(conn, http2.FrameSettings)
			if err != nil {
				return err
			}

			// Send WINDOW_UPDATE with increment size 2.
			conn.WriteWindowUpdate(streamID, 2)

			// Wait for DATA frame...
			passed := false
			for !conn.Closed {
				ev := conn.WaitEvent()

				switch event := ev.(type) {
				case spec.EventDataFrame:
					actual = event
					passed = (event.Header().Length == 1)
				case spec.EventTimeout:
					if actual == nil {
						actual = event
					}
				default:
					actual = ev
				}

				if passed {
					break
				}
			}

			if !passed {
				expected := []string{
					fmt.Sprintf("DATA Frame (length:1, flags:0x00, stream_id:%d)", streamID),
				}

				return &spec.TestError{
					Expected: expected,
					Actual:   actual.String(),
				}
			}

			return nil
		},
	})

	// An endpoint MUST treat a change to SETTINGS_INITIAL_WINDOW_SIZE
	// that causes any flow-control window to exceed the maximum size
	// as a connection error (Section 5.4.1) of type FLOW_CONTROL_ERROR.
	tg.AddTestCase(&spec.TestCase{
		Desc:        "Sends a SETTINGS_INITIAL_WINDOW_SIZE settings with an exceeded maximum window size value",
		Requirement: "The endpoint MUST treat this as a connection error of type FLOW_CONTROL_ERROR.",
		Run: func(c *config.Config, conn *spec.Conn) error {
			err := conn.Handshake()
			if err != nil {
				return err
			}

			// SETTINGS frame:
			// SETTINGS_INITIAL_WINDOW_SIZE: 2147483648
			conn.Send([]byte("\x00\x00\x06\x04\x00\x00\x00\x00\x00"))
			conn.Send([]byte("\x00\x04\x80\x00\x00\x00"))

			return spec.VerifyConnectionError(conn, http2.ErrCodeFlowControl)
		},
	})

	return tg
}
