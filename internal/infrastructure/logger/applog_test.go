package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func decodeLog(t *testing.T, buf *bytes.Buffer) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &m); err != nil {
		t.Fatalf("failed to decode log JSON: %v\nraw: %s", err, buf.String())
	}
	return m
}

func newBufLogger(opts ...StandardLoggerOption) (*StandardLogger, *bytes.Buffer) {
	var buf bytes.Buffer
	opts = append([]StandardLoggerOption{WithOutput(&buf)}, opts...)
	return NewStandardLogger(opts...), &buf
}

func TestNewStandardLogger_Info(t *testing.T) {
	l, buf := newBufLogger(WithAppID("go-nixcopy"), WithAppVersion("1.0.0"))
	l.Info("hello world")
	m := decodeLog(t, buf)

	if m["log_type"] != string(LogTypeApp) {
		t.Errorf("log_type = %v, want %v", m["log_type"], LogTypeApp)
	}
	if m["level"] != string(LogLevelInfo) {
		t.Errorf("level = %v, want %v", m["level"], LogLevelInfo)
	}
	if m["message"] != "hello world" {
		t.Errorf("message = %v, want hello world", m["message"])
	}
	if m["app_id"] != "go-nixcopy" {
		t.Errorf("app_id = %v, want go-nixcopy", m["app_id"])
	}
	if m["app_version"] != "1.0.0" {
		t.Errorf("app_version = %v, want 1.0.0", m["app_version"])
	}
	if _, ok := m["event_date_time"]; !ok {
		t.Error("event_date_time field missing")
	}
}

func TestNewStandardLogger_Warn(t *testing.T) {
	l, buf := newBufLogger()
	l.Warn("watch out")
	m := decodeLog(t, buf)
	if m["level"] != string(LogLevelWarn) {
		t.Errorf("level = %v, want WARN", m["level"])
	}
}

func TestNewStandardLogger_Error(t *testing.T) {
	l, buf := newBufLogger()
	l.Error("boom")
	m := decodeLog(t, buf)
	if m["level"] != string(LogLevelError) {
		t.Errorf("level = %v, want ERROR", m["level"])
	}
}

func TestNewStandardLogger_Debug(t *testing.T) {
	l, buf := newBufLogger()
	l.Debug("verbose")
	m := decodeLog(t, buf)
	if m["level"] != string(LogLevelDebug) {
		t.Errorf("level = %v, want DEBUG", m["level"])
	}
}

func TestStandardLogger_InfoReqEx(t *testing.T) {
	l, buf := newBufLogger()
	l.InfoReqEx("calling external service")
	m := decodeLog(t, buf)
	if m["log_type"] != string(LogTypeReqEx) {
		t.Errorf("log_type = %v, want %v", m["log_type"], LogTypeReqEx)
	}
	if m["level"] != string(LogLevelInfo) {
		t.Errorf("level = %v, want INFO", m["level"])
	}
}

func TestStandardLogger_InfoResEx(t *testing.T) {
	l, buf := newBufLogger()
	l.InfoResEx("response received")
	m := decodeLog(t, buf)
	if m["log_type"] != string(LogTypeResEx) {
		t.Errorf("log_type = %v, want %v", m["log_type"], LogTypeResEx)
	}
}

func TestStandardLogger_WarnResEx(t *testing.T) {
	l, buf := newBufLogger()
	l.WarnResEx("slow response")
	m := decodeLog(t, buf)
	if m["log_type"] != string(LogTypeResEx) {
		t.Errorf("log_type = %v, want %v", m["log_type"], LogTypeResEx)
	}
	if m["level"] != string(LogLevelWarn) {
		t.Errorf("level = %v, want WARN", m["level"])
	}
}

func TestStandardLogger_ErrorReqEx(t *testing.T) {
	l, buf := newBufLogger()
	l.ErrorReqEx("request failed")
	m := decodeLog(t, buf)
	if m["log_type"] != string(LogTypeReqEx) {
		t.Errorf("log_type = %v, want %v", m["log_type"], LogTypeReqEx)
	}
	if m["level"] != string(LogLevelError) {
		t.Errorf("level = %v, want ERROR", m["level"])
	}
}

func TestStandardLogger_ErrorResEx(t *testing.T) {
	l, buf := newBufLogger()
	l.ErrorResEx("bad response")
	m := decodeLog(t, buf)
	if m["log_type"] != string(LogTypeResEx) {
		t.Errorf("log_type = %v, want %v", m["log_type"], LogTypeResEx)
	}
	if m["level"] != string(LogLevelError) {
		t.Errorf("level = %v, want ERROR", m["level"])
	}
}

func TestStandardLogger_WithCorrelation(t *testing.T) {
	l, _ := newBufLogger()
	child, buf2 := func() (*StandardLogger, *bytes.Buffer) {
		var buf bytes.Buffer
		c := l.WithCorrelation("corr-123", "req-456")
		c.output = &buf
		return c, &buf
	}()
	child.Info("msg")
	m := decodeLog(t, buf2)
	if m["correlation_id"] != "corr-123" {
		t.Errorf("correlation_id = %v, want corr-123", m["correlation_id"])
	}
	if m["request_id"] != "req-456" {
		t.Errorf("request_id = %v, want req-456", m["request_id"])
	}
}

func TestStandardLogger_WithRequest(t *testing.T) {
	base, _ := newBufLogger()
	parent := base.WithCorrelation("corr-abc", "req-old")
	var buf bytes.Buffer
	child := parent.WithRequest("req-new")
	child.output = &buf
	child.Info("msg")
	m := decodeLog(t, &buf)
	if m["correlation_id"] != "corr-abc" {
		t.Errorf("correlation_id = %v, want corr-abc", m["correlation_id"])
	}
	if m["request_id"] != "req-new" {
		t.Errorf("request_id = %v, want req-new", m["request_id"])
	}
}

func TestStandardLogger_WithTrace(t *testing.T) {
	l, _ := newBufLogger()
	var buf bytes.Buffer
	child := l.WithTrace("trace-xyz", "span-123")
	child.output = &buf
	child.Info("msg")
	m := decodeLog(t, &buf)
	if m["trace_id"] != "trace-xyz" {
		t.Errorf("trace_id = %v, want trace-xyz", m["trace_id"])
	}
	if m["span_id"] != "span-123" {
		t.Errorf("span_id = %v, want span-123", m["span_id"])
	}
}

func TestStandardLogger_WithOptions(t *testing.T) {
	l, buf := newBufLogger(
		WithAppID("myapp"),
		WithAppVersion("2.0"),
		WithServiceID("svc-1"),
		WithPodName("pod-a"),
	)
	l.Info("opts")
	m := decodeLog(t, buf)
	checks := map[string]string{
		"app_id":           "myapp",
		"app_version":      "2.0",
		"service_id":       "svc-1",
		"service_pod_name": "pod-a",
	}
	for k, want := range checks {
		if m[k] != want {
			t.Errorf("%s = %v, want %v", k, m[k], want)
		}
	}
}

func TestNopLogger_NoOutput(t *testing.T) {
	l := NewNopLogger()
	// calling any method on nop must not panic and must produce no output
	l.Info("ignored")
	l.Warn("ignored")
	l.Error("ignored")
	l.Debug("ignored")
	l.InfoReqEx("ignored")
	l.InfoResEx("ignored")
	l.WarnResEx("ignored")
	l.ErrorReqEx("ignored")
	l.ErrorResEx("ignored")
}

func TestGenerateID_Format(t *testing.T) {
	id := GenerateID()
	parts := strings.Split(id, "-")
	if len(parts) != 5 {
		t.Fatalf("GenerateID = %q: want 5 parts separated by '-', got %d", id, len(parts))
	}
	lengths := []int{8, 4, 4, 4, 12}
	for i, p := range parts {
		if len(p) != lengths[i] {
			t.Errorf("part[%d] = %q, want length %d", i, p, lengths[i])
		}
	}
	// version nibble must be 4
	if id[14] != '4' {
		t.Errorf("version nibble = %c, want 4", id[14])
	}
}

func TestGenerateID_Unique(t *testing.T) {
	if GenerateID() == GenerateID() {
		t.Error("two consecutive GenerateID() calls returned the same value")
	}
}

func TestF_Field(t *testing.T) {
	f := F("key", "val")
	if f.key != "key" || f.value != "val" {
		t.Errorf("F() = {%q, %v}, want {key, val}", f.key, f.value)
	}
}

func TestFError_Field(t *testing.T) {
	err := errors.New("something broke")
	f := FError(err)
	if f.key != "error" {
		t.Errorf("FError key = %q, want error", f.key)
	}
	if f.value != "something broke" {
		t.Errorf("FError value = %v, want 'something broke'", f.value)
	}
}

func TestFDurationMs_Field(t *testing.T) {
	f := FDurationMs(250 * time.Millisecond)
	if f.key != "execution_time" {
		t.Errorf("FDurationMs key = %q, want execution_time", f.key)
	}
	if f.value != int64(250) {
		t.Errorf("FDurationMs value = %v, want 250", f.value)
	}
}

func TestStandardLogger_StandardFields(t *testing.T) {
	l, buf := newBufLogger()
	l.Info("msg",
		FDurationMs(100*time.Millisecond),
		F("caller_user", "alice"),
		F("caller_address", "10.0.0.1"),
		F("code_location", "main.go:42"),
		F("geo_location", "TH"),
	)
	m := decodeLog(t, buf)
	if m["execution_time"] != float64(100) {
		t.Errorf("execution_time = %v, want 100", m["execution_time"])
	}
	if m["caller_user"] != "alice" {
		t.Errorf("caller_user = %v, want alice", m["caller_user"])
	}
	if m["caller_address"] != "10.0.0.1" {
		t.Errorf("caller_address = %v, want 10.0.0.1", m["caller_address"])
	}
	if m["code_location"] != "main.go:42" {
		t.Errorf("code_location = %v, want main.go:42", m["code_location"])
	}
	if m["geo_location"] != "TH" {
		t.Errorf("geo_location = %v, want TH", m["geo_location"])
	}
}

func TestStandardLogger_ExtraFields(t *testing.T) {
	l, buf := newBufLogger()
	l.Info("msg", F("file_count", 42), F("backend", "s3"))
	m := decodeLog(t, buf)
	if m["file_count"] != float64(42) {
		t.Errorf("file_count = %v, want 42", m["file_count"])
	}
	if m["backend"] != "s3" {
		t.Errorf("backend = %v, want s3", m["backend"])
	}
}

func TestStandardLogger_ParentUnmutated(t *testing.T) {
	l, _ := newBufLogger()
	_ = l.WithCorrelation("c", "r")
	var buf bytes.Buffer
	l.output = &buf
	l.Info("msg")
	m := decodeLog(t, &buf)
	if _, ok := m["correlation_id"]; ok {
		t.Error("parent logger should not have correlation_id after WithCorrelation call")
	}
}
