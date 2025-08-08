package torrnado

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

const PARSE_TEST_DATA_FOLDER = "/home/roos/utv/tools/torrnado/testdata/parse"

func Test_parser_Parse(t *testing.T) {
	discardLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	pattern := "*.html" 
	matches, err := filepath.Glob(filepath.Join(PARSE_TEST_DATA_FOLDER, pattern))
	if err != nil {
		t.Fatalf("Error globbing files: %v", err)
	}

	if len(matches) == 0 {
		t.Fatalf("No files found matching the pattern")
	}

	type fields struct {
		log *slog.Logger
	}
	type args struct {
		html_source string
	}
	type test struct {
		name    string
		fields  fields
		args    args
		want    parsed_html
		wantErr bool
	}
	tests := []test{}

	for _, match := range matches {

		html_source_file := match
		want_json_file := filepath.Join(
			PARSE_TEST_DATA_FOLDER,
			fmt.Sprintf("%s.json", filepath.Base(html_source_file)),
		)

		html_source, err := os.ReadFile(html_source_file)
		if err != nil {
			t.Fatalf("error reading html_source: %v", err)
		}

		want_json, err := os.ReadFile(want_json_file)
		if err != nil {
			t.Fatalf("error reading wanted json: %v", err)
		}

		var want parsed_html
		err = json.Unmarshal(want_json, &want)
		if err != nil {
			t.Fatalf("error unmarshal wanted json: %v", err)
		}

		tests = append(tests, 
			test{"success", fields{discardLogger}, args{string(html_source)}, want, false},
		)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &parser{
				log: tt.fields.log,
			}
			got, err := p.Parse(tt.args.html_source)
			if (err != nil) != tt.wantErr {
				t.Errorf("parser.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parser.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
