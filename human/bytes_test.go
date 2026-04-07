package human

import (
	"math"
	"testing"
)

func TestParseByteSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    string
		want float64
	}{
		{name: "empty", s: "", want: 0},
		{name: "no number", s: "MiB", want: 0},
		{name: "bare number", s: "42", want: 42},
		{name: "bytes", s: "512 B", want: 512},
		{name: "bytes word", s: "1 bytes", want: 1},
		{name: "byte word", s: "1 byte", want: 1},
		{name: "KiB", s: "2.50 KiB", want: 2.5 * KiB},
		{name: "MiB", s: "27.61 MiB", want: 27.61 * MiB},
		{name: "GiB", s: "3.50 GiB", want: 3.5 * GiB},
		{name: "TiB", s: "1.25 TiB", want: 1.25 * TiB},
		{name: "PiB", s: "1 PiB", want: PiB},
		{name: "EiB", s: "1 EiB", want: EiB},
		{name: "KB", s: "1.50 KB", want: 1.5 * KB},
		{name: "kB", s: "1.50 kB", want: 1.5 * KB},
		{name: "MB", s: "10 MB", want: 10 * MB},
		{name: "GB", s: "2.5 GB", want: 2.5 * GB},
		{name: "TB", s: "1 TB", want: TB},
		{name: "PB", s: "1 PB", want: PB},
		{name: "EB", s: "1 EB", want: EB},
		{name: "whitespace", s: "  27.61 MiB  ", want: 27.61 * MiB},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := ParseByteSize(test.s)
			if math.Abs(got-test.want) > 0.01 {
				t.Fatalf("ParseByteSize(%q) = %v, want %v", test.s, got, test.want)
			}
		})
	}
}

func TestFormatIECBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		b    float64
		want string
	}{
		{name: "zero", b: 0, want: "0 B"},
		{name: "bytes", b: 512, want: "512 B"},
		{name: "one KiB", b: KiB, want: "1.00 KiB"},
		{name: "KiB", b: 2.5 * KiB, want: "2.50 KiB"},
		{name: "one MiB", b: MiB, want: "1.00 MiB"},
		{name: "MiB", b: 27.61 * MiB, want: "27.61 MiB"},
		{name: "one GiB", b: GiB, want: "1.00 GiB"},
		{name: "GiB", b: 3.5 * GiB, want: "3.50 GiB"},
		{name: "one TiB", b: TiB, want: "1.00 TiB"},
		{name: "TiB", b: 1.25 * TiB, want: "1.25 TiB"},
		{name: "one PiB", b: PiB, want: "1.00 PiB"},
		{name: "one EiB", b: EiB, want: "1.00 EiB"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := FormatIECBytes(test.b); got != test.want {
				t.Fatalf("FormatIECBytes(%v) = %q, want %q", test.b, got, test.want)
			}
		})
	}
}
