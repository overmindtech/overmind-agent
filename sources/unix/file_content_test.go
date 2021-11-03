package unix

import (
	"fmt"
	"path"
	"runtime"
	"testing"

	"github.com/overmindtech/overmind-agent/sources/util"
)

func TestFileContentGet(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	exampleTextFile := path.Join(path.Dir(filename), "test/example_file.txt")
	exampleBinaryFile := path.Join(path.Dir(filename), "test/example_file.png")

	t.Run("With a text file", func(t *testing.T) {
		src := FileContentSource{}

		item, err := src.Get(util.LocalContext, exampleTextFile)

		if err != nil {
			t.Fatal(err)
		}

		var contentInterface interface{}

		contentInterface, err = item.Attributes.Get("content_base64")

		if err != nil {
			t.Fatal(err)
		}

		expectedContent := `R290IHNvbWUgdGV4dCBpbm5pdA==`

		if fmt.Sprint(contentInterface) != expectedContent {
			t.Fatal("File content did not match expected")
		}
	})

	t.Run("With a binary file", func(t *testing.T) {
		src := FileContentSource{}

		item, err := src.Get(util.LocalContext, exampleBinaryFile)

		if err != nil {
			t.Fatal(err)
		}

		var contentInterface interface{}

		contentInterface, err = item.Attributes.Get("content_base64")

		if err != nil {
			t.Fatal(err)
		}

		if len(fmt.Sprint(contentInterface)) != 2352 {
			t.Fatalf("Expected content length to be 2352, got %v", len(fmt.Sprint(contentInterface)))
		}
	})

	t.Run("With a size limit", func(t *testing.T) {
		src := FileContentSource{
			MaxSize: 64,
		}

		_, err := src.Get(util.LocalContext, exampleTextFile)

		if err != nil {
			t.Fatal(err)
		}

		_, err = src.Get(util.LocalContext, exampleBinaryFile)

		if err == nil {
			t.Fatal("Expected error when reading file with size > 64B")
		}
	})
}
