package serial_test

import (
	"testing"
	"time"
	"github.com/jaracil/serial"

)

const (
	ttyPath = "/dev/ttyUSB1"
)

//////////////////////////////////////////////////////
// Tests
//////////////////////////////////////////////////////

// Tests are done using https://github.com/MihaiLupoiu/olimex-esp32-bit-banging.
func TestEchoNormalSerial(t *testing.T) {
	var tty *serial.Serial
	var err error

	defer func() {
		if tty != nil {
			tty.Close()
			tty = nil
		}
	}()

	if tty, err = serial.Open(ttyPath); err != nil {
		t.Errorf("serial: Error %s opening tty %s", err, ttyPath)
	}
	
	time.Sleep(time.Millisecond * 500)
	
	if err = tty.SetSpeed(9600); err != nil {
		t.Errorf("serial: Error %s setting bitrate in tty %s", err, ttyPath)
	}

	if err = tty.SetStopBits(int(2)); err != nil {
		t.Errorf("serial: Error %s setting stop bits in tty %s", err, ttyPath)
	}

	if err = tty.SetParity(serial.PAR_NONE); err != nil {
		t.Errorf("serial: Error %s setting parity bits in tty %s", err, ttyPath)
	}
	tty.Flush(serial.FLUSH_IO)

	
	msg := byte(24)
	if err = tty.WriteByte(msg); err != nil {
		t.Errorf("serial: Error %s writing: %v", err, msg)
	}

	if b, err := tty.ReadByte(); err == nil {
		if b != msg {
			t.Errorf("recv %d; want %d", b, msg)
		}
	} else {
		t.Errorf("Error reading from serial %s", err)
	}
}