package compiler

import (
	"fmt"
	"os"
	"os/exec"
)

// java -jar compiler.jar  --warning_level=VERBOSE --compilation_level=ADVANCED_OPTIMIZATIONS  --js_output_file=b/content_scripts.js compatible.js debug.js config.js debug.js  page-eater.js

func Compile(files []string, output_file string, warning_level, compilation_level string) error {
	args := []string{"-jar", "compiler.jar"}
	if len(warning_level) != 0 {
		args = append(args, fmt.Sprint("--warning_level=", warning_level))
	}

	if len(compilation_level) != 0 {
		args = append(args, fmt.Sprint("--compilation_level=", compilation_level))
	}

	args = append(args, fmt.Sprint("--js_output_file=", output_file))

	args = append(args, files...)

	cmd := exec.Command("java", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err := cmd.Start()
	if err != nil {
		return err
	}

	cmd.Wait()

	if cmd.ProcessState.Success() == false {
		return fmt.Errorf("执行错误。")
	}

	return nil
}
