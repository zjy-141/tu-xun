package logger

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

type StdWriter struct {
	*logrus.Logger
}

func (sw StdWriter) Write(info []byte) (n int, err error) {
	sw.Logger.Errorf("Find stderr: %s", string(info))
	return len(info), nil
}

// capture stderr to log file
func RedirectStderr(logWriter *StdWriter) {
	r, w, err := os.Pipe()
	if err != nil {
		logrus.Errorf("Failed to create pipe for stderr redirection: %v", err)
		return
	}
	os.Stderr = w

	go func() {
		_, err := io.Copy(logWriter, r)
		if err != nil {
			logrus.Errorf("Failed to copy stderr to log writer: %v", err)
		}
		err = r.Close()
		if err != nil {
			logrus.Errorf("Failed to close pipe reader: %v", err)
		}
	}()
}
