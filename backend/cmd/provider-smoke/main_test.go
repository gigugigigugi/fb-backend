package main

import "testing"

func TestParseChannels(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr bool
		email   bool
		sms     bool
	}{
		{name: "default both", input: "", email: true, sms: true},
		{name: "explicit both", input: "both", email: true, sms: true},
		{name: "email only", input: "email", email: true, sms: false},
		{name: "sms only", input: "sms", email: false, sms: true},
		{name: "trim and case", input: "  Email ", email: true, sms: false},
		{name: "invalid", input: "push", wantErr: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseChannels(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.email != tc.email || got.sms != tc.sms {
				t.Fatalf("unexpected channels: %+v", got)
			}
		})
	}
}
