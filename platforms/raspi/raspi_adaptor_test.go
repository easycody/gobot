package raspi

import (
	"testing"

	"github.com/hybridgroup/gobot"
	"github.com/hybridgroup/gobot/sysfs"
)

type NullReadWriteCloser struct{}

func (NullReadWriteCloser) Write(p []byte) (int, error) {
	return len(p), nil
}
func (NullReadWriteCloser) Read(b []byte) (int, error) {
	return len(b), nil
}
func (NullReadWriteCloser) Close() error {
	return nil
}

func initTestRaspiAdaptor() *RaspiAdaptor {
	readFile = func() ([]byte, error) {
		return []byte(`
Hardware        : BCM2708
Revision        : 0010
Serial          : 000000003bc748ea
`), nil
	}
	a := NewRaspiAdaptor("myAdaptor")
	a.Connect()
	return a
}

func TestRaspiAdaptor(t *testing.T) {
	readFile = func() ([]byte, error) {
		return []byte(`
Hardware        : BCM2708
Revision        : 0010
Serial          : 000000003bc748ea
`), nil
	}
	a := NewRaspiAdaptor("myAdaptor")
	gobot.Assert(t, a.Name(), "myAdaptor")
	gobot.Assert(t, a.i2cLocation, "/dev/i2c-1")
	gobot.Assert(t, a.revision, "3")

	readFile = func() ([]byte, error) {
		return []byte(`
Hardware        : BCM2708
Revision        : 000D
Serial          : 000000003bc748ea
`), nil
	}
	a = NewRaspiAdaptor("myAdaptor")
	gobot.Assert(t, a.i2cLocation, "/dev/i2c-1")
	gobot.Assert(t, a.revision, "2")

	readFile = func() ([]byte, error) {
		return []byte(`
Hardware        : BCM2708
Revision        : 0002
Serial          : 000000003bc748ea
`), nil
	}
	a = NewRaspiAdaptor("myAdaptor")
	gobot.Assert(t, a.i2cLocation, "/dev/i2c-0")
	gobot.Assert(t, a.revision, "1")

}
func TestRaspiAdaptorFinalize(t *testing.T) {
	a := initTestRaspiAdaptor()
	fs := sysfs.NewMockFilesystem([]string{
		"/sys/class/gpio/export",
		"/sys/class/gpio/unexport",
	})

	sysfs.SetFilesystem(fs)
	a.DigitalWrite("3", 1)
	a.i2cDevice = new(NullReadWriteCloser)
	gobot.Assert(t, len(a.Finalize()), 0)
}

func TestRaspiAdaptorDigitalIO(t *testing.T) {
	a := initTestRaspiAdaptor()
	fs := sysfs.NewMockFilesystem([]string{
		"/sys/class/gpio/export",
		"/sys/class/gpio/unexport",
		"/sys/class/gpio/gpio4/value",
		"/sys/class/gpio/gpio4/direction",
		"/sys/class/gpio/gpio27/value",
		"/sys/class/gpio/gpio27/direction",
	})

	sysfs.SetFilesystem(fs)

	a.DigitalWrite("7", 1)
	gobot.Assert(t, fs.Files["/sys/class/gpio/gpio4/value"].Contents, "1")

	a.DigitalWrite("13", 1)
	i, _ := a.DigitalRead("13")
	gobot.Assert(t, i, 1)
}

func TestRaspiAdaptorI2c(t *testing.T) {
	a := initTestRaspiAdaptor()
	fs := sysfs.NewMockFilesystem([]string{
		"/dev/i2c-1",
	})
	sysfs.SetFilesystem(fs)
	sysfs.SetSyscall(&sysfs.MockSyscall{})
	a.I2cStart(0xff)

	a.I2cWrite([]byte{0x00, 0x01})
	data, _ := a.I2cRead(2)
	gobot.Assert(t, data, []byte{0x00, 0x01})
}
