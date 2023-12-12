package hash

import (
	"testing"
)

func TestTicket(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   int64
		wantErr bool
	}{
		{
			name: "Valid input",
			args: args{
				input: "xx",
			},
			want:    "123456",
			want1:   15,
			wantErr: false,
		},
		{
			name: "Invalid input",
			args: args{
				input: "XYZ789",
			},
			want:    "",
			want1:   0,
			wantErr: true,
		},
		// Add more test cases if needed
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := Ticket(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ticket() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Ticket() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Ticket() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
