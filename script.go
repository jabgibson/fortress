package fortress

import (
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type ScriptOrder struct {
	Order
	SourceDir string `toml:"dir"`
	Shell     string `toml:"shell"`
	Source    string `toml:"source"`
	Output    string `toml:"output"`
}

func (s ScriptOrder) Sequence() int {
	return s.Seq
}

func (s ScriptOrder) ExecuteOrder(context OrderContext) (report Report) {
	// Note original environment variables and overwrite them with context given variables
	originalEnvs := map[string]string{}
	for k, v := range context.EnvVars {
		originalEnvs[k] = os.Getenv(k)
		if err := os.Setenv(k, v); err != nil {
			report.Errors = append(report.Errors, err)
			report.ExitCode = 1
			return
		}
	}

	// Guarantee environment variables are restored in case of early exit
	defer func() {
		for k, v := range originalEnvs {
			if v != "" {
				if err := os.Setenv(k, v); err != nil {
					report.Errors = append(report.Errors, err)
					report.ExitCode = 1
					return
				}
			}
		}
	}()

	// If shell is not defined, define it with a detault shell
	if s.Shell == "" {
		// Find the path for "sh", if it can't be found, default to /bin/sh
		path, err := exec.LookPath("sh")
		if err != nil {
			path = "/bin/sh"
		}
		s.Shell = path
	}

	// If source directory is not defined, set it to the systems temp directory
	if s.SourceDir == "" {
		s.SourceDir = os.TempDir()
	}

	// Build the temporary script name and full path
	tempScriptName := uuid.New().String()
	scriptFullPath := s.SourceDir + string(os.PathSeparator) + tempScriptName

	// Check script source for Fortress Shared Data expressions and replace if applicable
	source := s.Source
	for k, v := range context.Data {
		// Check for and replace Fortress Shared Data if exists
		if strings.HasPrefix(k, SharedDataPrefix) {
			source = strings.Replace(source, "#shared["+strings.TrimPrefix(k, SharedDataPrefix)+"]", v, -1)
		}
	}

	// Build the source into script
	scriptSource := fmt.Sprintf("#! %s\n\n%s\n", s.Shell, source)

	if err := ioutil.WriteFile(scriptFullPath, []byte(scriptSource), 0744); err != nil {
		report.Errors = append(report.Errors, err)
		report.ExitCode = 1
		return
	}
	cmd := exec.Command(scriptFullPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		report.Errors = append(report.Errors, err)
		report.ExitCode = 1
	}
	report.Output = output

	// Clean up temporary resources
	fmt.Println("debug:\n" + string(output)) //TODO debug
	return
}

func (s ScriptOrder) Self() Order {
	return s.Order
}
