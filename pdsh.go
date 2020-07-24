package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	fmt.Printf("Pdsh v0.0.3\nCopyright 2020 Tim_Paik @ Pd2 All rights reserved.\n")
	for {
		PrintPwd()
		input := bufio.NewReader(os.Stdin)
		if line, isPrefix, err := input.ReadLine(); err != nil {
			if err == io.EOF {
				os.Exit(1)
			}
			PrintError(err)
			return
		} else if isPrefix {
			PrintError(fmt.Errorf("The command is too long. "))
			continue
		} else {
			if len(line) == 0 {
				continue
			}
			cmd := strings.Fields(string(line))
			switch cmd[0] {
			case "cd":
				if err := os.Chdir(cmd[1]); err != nil {
					PrintError(err)
					continue
				}
			case "exit":
				os.Exit(0)
			default:
				execCmd := exec.Cmd{
					Path: cmd[0],
					Args: cmd,
				}
				if filepath.Base(cmd[0]) == cmd[0] {
					if lp, err := exec.LookPath(cmd[0]); err != nil {
						PrintError(fmt.Errorf("pdsh: command not found: %s", cmd[0]))
						continue
					} else {
						execCmd.Path = lp
					}
				}
				var stdout, stderr []byte
				var errStdout, errStderr error
				stdoutIn, _ := execCmd.StdoutPipe()
				stderrIn, _ := execCmd.StderrPipe()
				if err := execCmd.Start(); err != nil {
					PrintError(err)
					continue
				}
				go func() {
					stdout, errStdout = copyAndCapture(os.Stdout, stdoutIn)
				}()
				go func() {
					stderr, errStderr = copyAndCapture(os.Stderr, stderrIn)
				}()
				if err := execCmd.Wait(); err != nil {
					PrintError(err)
					continue
				}
				if errStdout != nil || errStderr != nil {
					PrintError(fmt.Errorf("failed to capture stdout or stderr\n"))
				}
			}
		}
	}
}

func PrintPwd() {
	if pwd, err := os.Getwd(); err != nil {
		PrintError(err)
		return
	} else {
		fmt.Printf("\n# " + pwd + "\n$ ")
	}
}

func PrintError(err error) {
	if _, err := fmt.Fprintln(os.Stderr, err); err != nil {
		println(err)
		return
	}
}

func copyAndCapture(w io.Writer, r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			if _, err := os.Stdout.Write(d); err != nil {
				return nil, err
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
}
