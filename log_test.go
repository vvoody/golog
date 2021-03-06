package golog

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	infoPath := filepath.Join(os.TempDir(), "test_info.log")
	debugPath := filepath.Join(os.TempDir(), "test_debug.log")
	os.Remove(infoPath)
	os.Remove(debugPath)

	infoWriter, err := NewFileWriter(infoPath)
	if err != nil {
		t.Error(err)
	}
	debugWriter, err := NewFileWriter(debugPath)
	if err != nil {
		t.Error(err)
	}

	infoHandler := NewHandler(InfoLevel, DefaultFormatter)
	infoHandler.AddWriter(infoWriter)
	debugHandler := &Handler{
		formatter: DefaultFormatter,
	}
	debugHandler.AddWriter(debugWriter)

	l := Logger{}
	l.AddHandler(infoHandler)
	l.AddHandler(debugHandler)

	l.Debugf("test %d", 1)

	stat, err := os.Stat(infoPath)
	if err != nil {
		t.Error(err)
	}
	if stat.Size() != 0 {
		t.Errorf("file size are %d", stat.Size())
	}

	debugContent, err := ioutil.ReadFile(debugPath)
	if err != nil {
		t.Error(err)
	}
	size1 := len(debugContent)
	if size1 == 0 {
		t.Error("debug log is empty")
	}

	l.Infof("test %d", 2)

	infoContent, err := ioutil.ReadFile(infoPath)
	if err != nil {
		t.Error(err)
	}

	parts := strings.Fields(string(infoContent))
	if len(parts) != 6 {
		t.Errorf("parts length are %d", len(parts))
	}
	if parts[0] != "[I" {
		t.Errorf("parts[0] is " + parts[0])
	}
	if len(parts[1]) != 10 {
		t.Errorf("parts[1] is " + parts[1])
	}
	if len(parts[2]) != 8 {
		t.Errorf("parts[2] is " + parts[2])
	}
	if !strings.HasPrefix(parts[3], "log_test:") {
		t.Errorf("parts[3] is " + parts[3])
	}
	if parts[4] != "test" {
		t.Errorf("parts[4] is " + parts[4])
	}
	if parts[5] != "2" {
		t.Errorf("parts[5] is " + parts[5])
	}

	debugContent, err = ioutil.ReadFile(debugPath)
	if err != nil {
		t.Error(err)
	}
	size2 := len(debugContent)
	if size2 != size1*2 {
		t.Errorf("debug log size are %d bytes", size2)
	}

	if !bytes.Equal(debugContent[size1:], infoContent) {
		t.Error("log contents are not equal")
	}
}

func BenchmarkBufferedFileLogger(b *testing.B) {
	path := filepath.Join(os.TempDir(), "test.log")
	os.Remove(path)
	w, err := NewBufferedFileWriter(path)
	if err != nil {
		b.Error(err)
		return
	}
	h := NewHandler(InfoLevel, DefaultFormatter)
	h.AddWriter(w)
	l := Logger{}
	l.AddHandler(h)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Infof("test")
		}
	})
	l.Close()
}

type discardWriter struct {
	io.Writer
}

func (w *discardWriter) Close() error {
	w.Writer = nil
	return nil
}

func BenchmarkDiscardLogger(b *testing.B) {
	w := &discardWriter{
		Writer: ioutil.Discard,
	}
	h := NewHandler(InfoLevel, DefaultFormatter)
	h.AddWriter(w)
	l := Logger{}
	l.AddHandler(h)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Infof("test")
		}
	})
	l.Close()
}
