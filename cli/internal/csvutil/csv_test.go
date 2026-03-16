package csvutil_test

import (
	"testing"

	"github.com/gechr/clib/cli/internal/csvutil"
	"github.com/stretchr/testify/require"
)

func TestAppend(t *testing.T) {
	tests := []struct {
		name string
		dst  []string
		raw  string
		want []string
	}{
		{
			name: "empty string",
			raw:  "",
			want: nil,
		},
		{
			name: "only commas",
			raw:  ",,",
			want: nil,
		},
		{
			name: "whitespace-only parts",
			raw:  " , , ",
			want: nil,
		},
		{
			name: "nil dst with single value",
			raw:  "foo",
			want: []string{"foo"},
		},
		{
			name: "multiple values",
			raw:  "a, b, c",
			want: []string{"a", "b", "c"},
		},
		{
			name: "values with extra whitespace",
			raw:  " foo , bar ",
			want: []string{"foo", "bar"},
		},
		{
			name: "appends to existing dst",
			dst:  []string{"x"},
			raw:  "a,b",
			want: []string{"x", "a", "b"},
		},
		{
			name: "mixed empty and valid",
			raw:  "a,,b, ,c",
			want: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := csvutil.Append(tt.dst, tt.raw)
			require.Equal(t, tt.want, got)
		})
	}
}
