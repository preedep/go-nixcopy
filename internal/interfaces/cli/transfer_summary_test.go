package cli

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestTransferSummary_JSONFields(t *testing.T) {
	s := transferSummary{
		Event:            "transfer_summary",
		TotalFiles:       3,
		Successful:       2,
		Skipped:          1,
		Failed:           0,
		BytesTransferred: 1048576,
		DurationMs:       500,
		AverageSpeedMBps: 2.0,
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(s); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	checks := map[string]any{
		"event":             "transfer_summary",
		"total_files":       float64(3),
		"successful":        float64(2),
		"skipped":           float64(1),
		"failed":            float64(0),
		"bytes_transferred": float64(1048576),
		"duration_ms":       float64(500),
		"average_speed_mbps": float64(2.0),
	}
	for key, want := range checks {
		got, ok := out[key]
		if !ok {
			t.Errorf("JSON missing field %q", key)
			continue
		}
		if got != want {
			t.Errorf("JSON[%q] = %v, want %v", key, got, want)
		}
	}

	// failed_files should be absent when nil (omitempty)
	if _, ok := out["failed_files"]; ok {
		t.Error("failed_files should be omitted when nil")
	}
}

func TestTransferSummary_AverageSpeedOmittedWhenZero(t *testing.T) {
	s := transferSummary{
		Event:      "transfer_summary",
		TotalFiles: 1,
		Successful: 1,
		DurationMs: 100,
		// AverageSpeedMBps left at zero
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(s); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	var out map[string]any
	_ = json.Unmarshal(buf.Bytes(), &out)

	if _, ok := out["average_speed_mbps"]; ok {
		t.Error("average_speed_mbps should be omitted when zero (omitempty)")
	}
}

func TestTransferSummary_FailedFiles(t *testing.T) {
	s := transferSummary{
		Event:      "transfer_summary",
		TotalFiles: 2,
		Successful: 1,
		Failed:     1,
		FailedFiles: []failedFile{
			{Path: "/src/bad.txt", Error: "connection reset"},
		},
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(s); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	raw, ok := out["failed_files"]
	if !ok {
		t.Fatal("failed_files missing from JSON")
	}
	arr, ok := raw.([]any)
	if !ok || len(arr) != 1 {
		t.Fatalf("failed_files: want array of 1, got %v", raw)
	}
	entry, ok := arr[0].(map[string]any)
	if !ok {
		t.Fatal("failed_files[0] is not an object")
	}
	if entry["path"] != "/src/bad.txt" {
		t.Errorf("failed_files[0].path = %v, want /src/bad.txt", entry["path"])
	}
	if entry["error"] != "connection reset" {
		t.Errorf("failed_files[0].error = %v, want 'connection reset'", entry["error"])
	}
}

func TestTransferSummary_EventAlwaysPresent(t *testing.T) {
	s := transferSummary{Event: "transfer_summary"}
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var out map[string]any
	_ = json.Unmarshal(b, &out)

	if out["event"] != "transfer_summary" {
		t.Errorf("event = %v, want transfer_summary", out["event"])
	}
}
