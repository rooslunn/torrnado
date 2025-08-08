package main

import (
	"log/slog"
	"testing"
)

func Test_parserCmd_parse(t *testing.T) {
	type fields struct {
		log *slog.Logger
	}
	type args struct {
		args []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := parserCmd{
				log: tt.fields.log,
			}
			if err := c.parse(tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("parserCmd.parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
