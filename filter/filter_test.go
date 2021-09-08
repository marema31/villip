package filter

import (
	"testing"
)

func TestFilter_Kind(t *testing.T) {
	type fields struct {
		kind Type
	}
	tests := []struct {
		name   string
		fields fields
		want   Type
	}{
		{
			"http",
			fields{HTTP},
			HTTP,
		},
		{
			"tcp",
			fields{TCP},
			TCP,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Filter{
				kind: tt.fields.kind,
			}
			if got := f.Kind(); got != tt.want {
				t.Errorf("Filter.Kind() = %v, want %v", got, tt.want)
			}
		})
	}
}
