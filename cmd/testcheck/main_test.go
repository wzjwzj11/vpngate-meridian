package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// generateAliveYAML used to emit each proxy's fields in Go map iteration order,
// which is randomized. Every regeneration therefore rewrote the whole file and
// broke anyone parsing the blocks positionally. The key order must be stable.
func TestGenerateAliveYAMLKeyOrderIsDeterministic(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "in.yaml")
	output := filepath.Join(dir, "out.yaml")

	// Keys deliberately written in a non-canonical order.
	in := `proxies:
  - udp: true
    server: "1.2.3.4"
    port: 443
    auth: "SHA1"
    name: "US test"
    type: openvpn
    proto: "tcp"
`
	if err := os.WriteFile(input, []byte(in), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	alive := map[string]int{"US test": 123}
	if err := generateAliveYAML(input, output, alive, 0); err != nil {
		t.Fatalf("generateAliveYAML: %v", err)
	}

	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	want := `proxies:
  # latency: 123ms
  - name: "US test"
    type: "openvpn"
    server: "1.2.3.4"
    port: 443
    proto: "tcp"
    auth: "SHA1"
    udp: true
`
	if string(got) != want {
		t.Fatalf("unexpected output\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// Repeated runs over the same input must produce byte-identical output.
func TestGenerateAliveYAMLIsReproducible(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "in.yaml")
	in := `proxies:
  - name: "A"
    type: openvpn
    server: "1.1.1.1"
    port: 443
    proto: "tcp"
    cipher: "AES-128-CBC"
    auth: "SHA1"
    udp: true
  - name: "B"
    type: openvpn
    server: "2.2.2.2"
    port: 1194
    proto: "udp"
    cipher: "AES-128-CBC"
    auth: "SHA1"
    udp: true
`
	if err := os.WriteFile(input, []byte(in), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	alive := map[string]int{"A": 100, "B": 200}

	var first string
	for i := 0; i < 20; i++ {
		output := filepath.Join(dir, "out.yaml")
		if err := generateAliveYAML(input, output, alive, 0); err != nil {
			t.Fatalf("run %d: %v", i, err)
		}
		got, err := os.ReadFile(output)
		if err != nil {
			t.Fatalf("run %d read: %v", i, err)
		}
		if i == 0 {
			first = string(got)
			continue
		}
		if string(got) != first {
			t.Fatalf("run %d differs from run 0\n--- run 0 ---\n%s\n--- run %d ---\n%s", i, first, i, got)
		}
	}

	if !strings.Contains(first, `  - name: "A"`) || !strings.Contains(first, `  - name: "B"`) {
		t.Fatalf("expected both proxies in output:\n%s", first)
	}
}
