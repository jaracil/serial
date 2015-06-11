package serial

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/jaracil/poll"
)

type Serial struct {
	f *poll.File
	//Characters ignored in LineRead
	LineIgnore string
	//Characters signaling end of line
	LineEnd string
}

const (
	PAR_NONE = iota // No parity
	PAR_EVEN        // Even parity
	PAR_ODD         // Odd parity
)

const (
	FLUSH_I  = iota // Flush input buffer
	FLUSH_O         // Flush output buffer
	FLUSH_IO        // Flush input/output buffers
)

var ErrTimeout = poll.ErrTimeout
var ErrClosed = poll.ErrClosed

// Open opens serial with default params.
//   Params:
//     path: Device path (Ex. "/dev/ttyUSB0")
//	 Default: 9600 8N1, soft/hard, flow controll off.
func Open(path string) (*Serial, error) {
	fd, err := open(path)
	if err != nil {
		return nil, err
	}
	pfd, err := poll.NewFile(uintptr(fd), path)
	if err != nil {
		return nil, err
	}
	s := &Serial{f: pfd, LineIgnore: "\r", LineEnd: "\n"}
	err = s.init()
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Close closes serial.
func (s *Serial) Close() error {
	err := s.f.Close()
	return err
}

// Read reads slice from serial.
func (s *Serial) Read(b []byte) (int, error) {
	return s.f.Read(b)
}

// WriteString writes string to serial.
func (s *Serial) WriteString(str string) (int, error) {
	return s.f.Write([]byte(str))
}

// Write writes byte slice to serial.
func (s *Serial) Write(b []byte) (int, error) {
	return s.f.Write(b)
}

// WriteByte writes one byte to serial.
func (s *Serial) WriteByte(c byte) error {
	_, e := s.f.Write([]byte{c})
	return e
}

// ReadByte reads one byte from serial.
func (s *Serial) ReadByte() (byte, error) {
	buf := make([]byte, 1)
	n, e := s.f.Read(buf)
	if n == 1 {
		return buf[0], nil
	}
	return 0, e
}

// Name returns serial file name.
func (s *Serial) Name() string {
	return s.f.Name()
}

// File returns serial os.File struct.
func (s *Serial) File() *poll.File {
	return s.f
}

// Fd returns serial file descriptor.
func (s *Serial) Fd() uintptr {
	return s.f.Fd()
}

// SetBits sets frame bits (5,6,7,8).
func (s *Serial) SetBits(bits int) error {
	return s.setBits(bits)
}

// SetSpeed sets serial speed.
func (s *Serial) SetSpeed(speed int) error {
	return s.setSpeed(speed)
}

// SetHwFlowCtrl enable or disable Hardware flow control.
func (s *Serial) SetHwFlowCtrl(hw bool) error {
	return s.setHwFlowCtrl(hw)
}

// SetSwFlowCtrl enable or disable software flow control.
func (s *Serial) SetSwFlowCtrl(sw bool) error {
	return s.setSwFlowCtrl(sw)
}

// SetStopBits sets stop bits, valid values are 1 or 2.
func (s *Serial) SetStopBits(stop int) error {
	switch stop {
	case 1:
		return s.setStopBits2(false)
	case 2:
		return s.setStopBits2(true)
	default:
		return errors.New("Invalid stop bits number")
	}
}

// SetParity sets parity mode:
//   PAR_NONE
//   PAR_EVEN
//   PAR_ODD
func (s *Serial) SetParity(mode int) error {
	return s.setParity(mode)
}

// SetLocal sets local mode. In local mode, modem control lines are ignored.
func (s *Serial) SetLocal(local bool) error {
	return s.setLocal(local)
}

// GetAttr sets Termios structure from serial attributes.
func (s *Serial) GetAttr(attr *Termios) error {
	return s.tcGetAttr(attr)
}

// SetAttr sets serial attributes from Termios structure.
func (s *Serial) SetAttr(attr *Termios) error {
	return s.tcSetAttr(attr)
}

// SetHub sets hangup mode (false -> don't reset DTR/RTS on exit).
func (s *Serial) SetHup(hup bool) error {
	return s.setHup(hup)
}

// InpWaiting returns number of bytes waiting on input buffer.
func (s *Serial) InpWaiting() (int, error) {
	return s.inpWaiting()
}

// OutWaiting returns number of bytes waiting on output buffer.
func (s *Serial) OutWaiting() (int, error) {
	return s.outWaiting()
}

// SetDeadline sets read/write deadline time
func (s *Serial) SetDeadline(t time.Time) error {
	return s.f.SetDeadline(t)
}

// SetReadDeadline sets read deadline time
func (s *Serial) SetReadDeadline(t time.Time) error {
	return s.f.SetReadDeadline(t)
}

// SetWriteDeadline sets write deadline time
func (s *Serial) SetWriteDeadline(t time.Time) error {
	return s.f.SetWriteDeadline(t)
}

// Flush buffers selected by mode:
//   FLUSH_I  input buffer
//   FLUSH_O  output buffer
//   FLUSH_IO input and output buffers
func (s *Serial) Flush(mode int) error {
	return s.flush(mode)
}

// SetCtrlBit sets level of modem control signal (DTR, RTS, ...)
func (s *Serial) SetCtrlBit(ctr int, level bool) error {
	return s.setCtrlBit(ctr, level)
}

// GetCtrl gets modem control bits
func (s *Serial) GetCtrl() (int, error) {
	return s.getCtrl()
}

// SetCtrl sets modem control bits
func (s *Serial) SetCtrl(ctr int) error {
	return s.setCtrl(ctr)
}

// ReadLine reads text line.
// Serial.LineIgnore field has characters to be ignored (by default "\r").
// Serial.LineEnd field has end of line characters (by default "\n").
func (s *Serial) ReadLine() (res string, err error) {
	var b byte
	for {
		if b, err = s.ReadByte(); err != nil {
			return
		}
		ch := string(b)
		if strings.Contains(s.LineIgnore, ch) {
			continue
		}
		if strings.Contains(s.LineEnd, ch) {
			break
		}
		res += ch
	}
	return
}

// WaitForRe reads lines from serial and waits for line matching one regular expresion from rexp slice.
// It returns the index of rexp slice matching text line, text line itself and error != nil on timeout or I/O error.
func (s *Serial) WaitForRe(rexp []string) (int, string, error) {
	var match string
	var err error

	for {
		if match, err = s.ReadLine(); err != nil {
			return -1, "", err
		}
		for i, re := range rexp {
			ok, err := regexp.MatchString(re, match)
			if err != nil {
				return -1, "", err
			}
			if ok {
				return i, match, nil
			}
		}
	}
}
