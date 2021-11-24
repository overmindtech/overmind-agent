package parser

import (
	"context"
	"os"
	"path"
	"regexp"
	"runtime"
	"testing"
)

var parserLabels = Parser{
	Ignore:  regexp.MustCompile(`^\s*#`),
	Capture: regexp.MustCompile(`^\s*(?P<food>[^,]+)\s*,\s*(?P<drink>[^,]+)\s*,\s*(?P<poo>[^,]+)\s*$`),
}

func TestParseWithLabels(t *testing.T) {
	var m []map[string]string

	_, filename, _, _ := runtime.Caller(0)
	exampleFile := path.Join(path.Dir(filename), "test/parser_test")

	file, err := os.Open(exampleFile)

	if err != nil {
		t.Error(err)
	}

	defer file.Close()

	m, err = parserLabels.Parse(context.Background(), file)

	if err != nil {
		t.Error(err)
	}

	if len(m) != 1 {
		t.Fatalf("Expected 1 match, got %v", len(m))
	}

	if m[0]["food"] != "🍤" {
		t.Errorf("Parsed value for food was %v expected %v", m[0]["food"], "🍤")
	}

	if m[0]["drink"] != "🍻" {
		t.Errorf("Parsed value for drink was %v expected %v", m[0]["drink"], "🍻")
	}

	if m[0]["poo"] != "💩" {
		t.Errorf("Parsed value for poo was %v expected %v", m[0]["poo"], "💩")
	}

}
