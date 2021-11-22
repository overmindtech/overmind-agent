package command

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/overmindtech/discovery"
)

// sleepCMD Returns a command and args that sleeps for a given period, depending
// on the GOOS
func sleepCMD(seconds int) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("ping 127.0.0.1 -n %v", seconds+1)
	}

	return fmt.Sprintf("sleep %v", seconds)
}

func TestRun(t *testing.T) {
	t.Run("with working comand", func(t *testing.T) {
		params := CommandParams{
			Command:      "hostname",
			ExpectedExit: 0,
			Timeout:      1 * time.Second,
		}

		item, err := params.Run()

		if err != nil {
			t.Fatal(err)
		}

		discovery.TestValidateItem(t, item)

		var name interface{}
		var exitCode interface{}
		var stdout interface{}
		var hostname string

		name, err = item.Attributes.Get("name")

		if err != nil {
			t.Error(err)
		}

		if name != "hostname" {
			t.Errorf("expected name to be \"hostname\" got %v", name)
		}

		exitCode, err = item.Attributes.Get("exitCode")

		if err != nil {
			t.Error(err)
		}

		if exitCode != float64(0) {
			t.Errorf("expected exitCode to be 0 got %v", exitCode)
		}

		stdout, err = item.Attributes.Get("stdout")

		if err != nil {
			t.Error(err)
		}

		hostname, err = os.Hostname()

		if err != nil {
			t.Error(err)
		}

		if fmt.Sprint(stdout) != hostname {
			t.Errorf("expected stdout to be %v got %v", hostname, stdout)
		}
	})

	t.Run("with timeout", func(t *testing.T) {
		t.Run("returning before the timeout", func(t *testing.T) {
			command := sleepCMD(1)

			params := CommandParams{
				Command: command,
				Timeout: 2 * time.Second,
			}

			item, err := params.Run()

			if err != nil {
				t.Error(err)
			}

			discovery.TestValidateItem(t, item)
		})

		t.Run("timing out", func(t *testing.T) {
			command := sleepCMD(10)
			timeoutrror := regexp.MustCompile("timed out")

			params := CommandParams{
				Command: command,
				Timeout: 500 * time.Millisecond,
			}

			_, err := params.Run()

			if err == nil || !timeoutrror.MatchString(err.Error()) {
				t.Error("No error returned or error was not timeout, command should have timed out")
			}
		})
	})

	t.Run("with non-zero exit codes", func(t *testing.T) {
		var command string

		if runtime.GOOS == "windows" {
			command = "ping somethingNotReal"
		} else {
			command = "cat somethingNotReal"
		}

		t.Run("an unexpected non-zero exit should fail", func(t *testing.T) {
			params := CommandParams{
				Command:      command,
				ExpectedExit: 0,
			}

			_, err := params.Run()

			if err == nil {
				t.Error("Expected command to fail but it didn't")
			}
		})

		t.Run("an expected non-zero exit should pass", func(t *testing.T) {
			params := CommandParams{
				Command:      command,
				ExpectedExit: 1,
			}

			_, err := params.Run()

			if err != nil {
				t.Fatal(err)
			}
		})
	})

	t.Run("shell builtins should work", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Not supported on windows")
		}

		params := CommandParams{
			Command: "exit 1",
		}

		_, err := params.Run()

		if err == nil {
			t.Fatal("expected command to fail but didn't")
		}
	})

	t.Run("stdout should work", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Not supported on windows")
		}

		params := CommandParams{
			Command: "echo qwerty",
		}

		item, err := params.Run()

		if err != nil {
			t.Fatal(err)
		}

		if stdout, _ := item.Attributes.Get("stdout"); stdout != "qwerty" {
			t.Errorf("expected stdout to be qwerty, got %v", stdout)
		}
	})

	t.Run("stderr should work", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Not supported on windows")
		}

		params := CommandParams{
			Command: "perl -e 'print STDERR qwerty'",
		}

		item, err := params.Run()

		if err != nil {
			t.Fatal(err)
		}

		if stderr, _ := item.Attributes.Get("stderr"); stderr != "qwerty" {
			t.Errorf("expected stderr to be qwerty, got \"%v\"", stderr)
		}
	})

	t.Run("stdout should work (windows)", func(t *testing.T) {
		if runtime.GOOS != "windows" {
			t.Skip("Only runs on windows")
		}

		params := CommandParams{
			Command:      "xcopy.exe",
			ExpectedExit: 4,
		}

		item, err := params.Run()

		if err != nil {
			t.Fatal(err)
		}

		if stdout, _ := item.Attributes.Get("stdout"); stdout != "0 File(s) copied" {
			t.Errorf("expected stdout to be \"0 File(s) copied\", got %v", stdout)
		}
	})

	t.Run("stderr should work (windows", func(t *testing.T) {
		if runtime.GOOS != "windows" {
			t.Skip("Only runs on windows")
		}

		params := CommandParams{
			Command:      "xcopy.exe",
			ExpectedExit: 4,
		}

		item, err := params.Run()

		if err != nil {
			t.Fatal(err)
		}

		if stderr, _ := item.Attributes.Get("stderr"); stderr != "Invalid number of parameters" {
			t.Errorf("expected stderr to be \"Invalid number of parameters\", got \"%v\"", stderr)
		}
	})
}

const jsonString = `{
	"timeout": "500ms",
	"stdin": "eWVzCmZvbyBiYXI=",
	"command": "cat hosts",
	"expected_exit": 0,
	"dir": "/etc",
	"env": {
		"TEST": "foo"
	}
}`

var jsonObject = CommandParams{
	Command:      "cat hosts",
	ExpectedExit: 0,
	Timeout:      500 * time.Millisecond,
	Dir:          "/etc",
	Env: map[string]string{
		"TEST": "foo",
	},
	STDIN: []byte{121, 101, 115, 10, 102, 111, 111, 32, 98, 97, 114},
}

func TestMarshalJSON(t *testing.T) {
	var b []byte
	var err error
	var resultString string

	b, err = json.MarshalIndent(jsonObject, "", "	")

	if err != nil {
		t.Fatal(err)
	}

	resultString = string(b)

	if resultString != jsonString {
		t.Errorf("JSON did not match. Expected:\n%v\nGot:\n%v", jsonString, resultString)
	}
}

func TestUnmarshalJSON(t *testing.T) {
	var cp CommandParams

	err := json.Unmarshal([]byte(jsonString), &cp)

	if err != nil {
		t.Fatal(err)
	}

	if cp.Command != jsonObject.Command {
		t.Errorf("Command did not match, got %v, expected %v", cp.Command, jsonObject.Command)
	}

	if cp.Dir != jsonObject.Dir {
		t.Errorf("Dir did not match, got %v, expected %v", cp.Dir, jsonObject.Dir)
	}

	if cp.ExpectedExit != jsonObject.ExpectedExit {
		t.Errorf("ExpectedExit did not match, got %v, expected %v", cp.ExpectedExit, jsonObject.ExpectedExit)
	}

	if cp.Timeout.String() != jsonObject.Timeout.String() {
		t.Errorf("Timeout.String() did not match, got %v, expected %v", cp.Timeout.String(), jsonObject.Timeout.String())
	}

	if cp.Env["TEST"] != jsonObject.Env["TEST"] {
		t.Errorf("Env[\"TEST\"] did not match, got %v, expected %v", cp.Env["TEST"], jsonObject.Env["TEST"])
	}

	if string(cp.STDIN) != "yes\nfoo bar" {
		t.Errorf("Expected stdin to  be the string \"yes\\nfoo bar\", got %v", string(cp.STDIN))
	}
}

func TestShellWrap(t *testing.T) {
	t.Run("with basic command", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping tests on windows")
		}

		cp := CommandParams{
			Command: "hostname",
		}

		_, newArgs, err := cp.ShellWrap()

		if err != nil {
			t.Fatal(err)
		}

		if newArgs[1] != "hostname" {
			t.Errorf("Expected wrapped command to be hostname, got %v", newArgs[1])
		}
	})

	t.Run("with a file with spaces", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping tests on windows")
		}

		cp := CommandParams{
			Command: "cat '/home/dylan/my file.txt'",
		}

		_, newArgs, err := cp.ShellWrap()

		if err != nil {
			t.Fatal(err)
		}

		if newArgs[1] != "cat '/home/dylan/my file.txt'" {
			t.Errorf("Expected wrapped command to be \"cat '/home/dylan/my file.txt'\", got %v", newArgs[1])
		}
	})
}

func TestPowershellWrap(t *testing.T) {
	t.Run("simple command", func(t *testing.T) {
		cp := CommandParams{
			Command: `Write-Host Hello!`,
		}
		_, args, err := cp.PowerShellWrap()

		if err != nil {
			t.Fatal(err)
		}

		expected := regexp.MustCompile("Write-Host Hello!")

		if !expected.MatchString(strings.Join(args, " ")) {
			t.Fatal("Expected to match 'Write-Host Hello!'")
		}
	})
}

func TestRunComplex(t *testing.T) {
	t.Run("exit code logic", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("bash not supproted on windows")
		}

		// In here I want to be testing complex stuff like if someone wanted pass a
		// whole shell script with logic, or even just a command that did something
		// based on the exit of something else using %% of || or something
		cp := CommandParams{
			Command: `[ -f /etc/foobar ] && echo "exists" || echo "does not exist"`,
		}

		item, err := cp.Run()

		if err != nil {
			t.Fatal(err)
		}

		if out, _ := item.Attributes.Get("stdout"); out != "does not exist" {
			t.Errorf("expected stdout to be 'does not exist', got %v", out)
		}
	})

	t.Run("full script", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("bash not supproted on windows")
		}

		_, filename, _, _ := runtime.Caller(0)
		scriptPath := path.Join(path.Dir(filename), "test/line_count.bash")
		content, err := os.ReadFile(scriptPath)

		if err != nil {
			t.Fatal(err)
		}

		cp := CommandParams{
			Command: string(content),
		}

		item, err := cp.Run()

		if err != nil {
			t.Fatal(err)
		}

		expected := regexp.MustCompile(`\/etc\/passwd:\s+\d+`)
		stdout := fmt.Sprint(item.Attributes.Get("stdout"))

		if !expected.MatchString(stdout) {
			t.Errorf("expected stdout to match %v, got %v", expected, stdout)
		}
	})
}

func TestRunComplexPowershell(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("non-windows not supproted")
	}

	_, filename, _, _ := runtime.Caller(0)
	scriptPath := path.Join(path.Dir(filename), "test/info.ps1")
	content, err := os.ReadFile(scriptPath)

	if err != nil {
		t.Fatal(err)
	}

	cp := CommandParams{
		Command: string(content),
	}

	_, err = cp.Run()

	if err != nil {
		t.Fatal(err)
	}
}

func TestSTDIN(t *testing.T) {
	t.Run("unix", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("bash not supproted on windows")
		}

		_, filename, _, _ := runtime.Caller(0)
		scriptPath := path.Join(path.Dir(filename), "test/password_prompt.bash")
		content, err := os.ReadFile(scriptPath)

		if err != nil {
			t.Fatal(err)
		}

		cp := CommandParams{
			Command: string(content),
			STDIN:   []byte("testing\n"),
			Timeout: 5 * time.Second,
		}

		item, err := cp.Run()

		if err != nil {
			t.Fatal(err)
		}

		discovery.TestValidateItem(t, item)

		if stdout, err := item.Attributes.Get("stdout"); err == nil {
			if match, err := regexp.MatchString("testing", fmt.Sprint(stdout)); !match || err != nil {
				t.Errorf("Expected stdout to be \"testing\", got \"%v\"\nError: %v", stdout, err)
			}
		} else {
			t.Error(err)
		}

	})
}
