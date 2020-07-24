package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var (
	Username string
	Hostname string
)

//非Windows编译请注释下面的var和winColorPrint里的proc，以及上面import的"syscall"
var (
	kernel32 *syscall.LazyDLL  = syscall.NewLazyDLL(`kernel32.dll`)
	proc     *syscall.LazyProc = kernel32.NewProc(`SetConsoleTextAttribute`)
)

func winColorPrint(s string, i int) {
	proc.Call(uintptr(syscall.Stdout), uintptr(i))
	fmt.Print(s)
	proc.Call(uintptr(syscall.Stdout), uintptr(7))
}

func init() {
	if userInfo, err := user.Current(); err != nil {
		os.Exit(2)
		return
	} else {
		switch runtime.GOOS {
		case "windows":
			Username = userInfo.Name
		default:
			Username = userInfo.Username
		}
	}
	if hostname, err := os.Hostname(); err != nil {
		return
	} else {
		Hostname = hostname
	}
}

func main() {
	fmt.Printf("Pdsh v0.0.4\nCopyright 2020 Tim_Paik @ Pd2 All rights reserved.\n")
	input := bufio.NewReader(os.Stdin)
	for {
		PrintPwd()
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
				var /*errStdin,*/ errStdout, errStderr error
				//stdinIn, _ := execCmd.StdinPipe()
				stdoutIn, _ := execCmd.StdoutPipe()
				stderrIn, _ := execCmd.StderrPipe()
				if err := execCmd.Start(); err != nil {
					PrintError(err)
					continue
				}
				isClose := false
				/*
					go func() {
						errStdin = Write(os.Stdin, stdinIn, &isClose)
					}()
				*/
				go func() {
					errStdout = Read(stdoutIn, os.Stdout, &isClose)
				}()
				go func() {
					errStderr = Read(stderrIn, os.Stderr, &isClose)
				}()
				if err := execCmd.Wait(); err != nil {
					isClose = true
					PrintError(err)
					continue
				}
				isClose = true
				if errStdout != nil || errStderr != nil /*|| errStdin != nil*/ {
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
		//fmt.Printf("\n# " + pwd + "\n$ ")
		switch runtime.GOOS {
		case "windows":
			fmt.Printf("\n")
			winColorPrint("# ", 1)
			fmt.Printf("%s @ ", Username)
			winColorPrint(Hostname, 2)
			fmt.Printf(" in %s [%s]\n", pwd, time.Now().Format("15:04:05"))
			winColorPrint("$ ", 4)
		default:
			//fmt.Printf(pwd)
			fmt.Printf("\n\033[34m# \033[0m%s @ \033[32m%s\033[0m in \033[33m%s\033[0m [%s]\n\033[31m$\033[0m ", Username, Hostname, pwd, time.Now().Format("15:04:05"))
		}
	}
}

func PrintError(err error) {
	if _, err := fmt.Fprintln(os.Stderr, err); err != nil {
		println(err)
		return
	}
}

func Read(from io.Reader, to io.Writer, isClose *bool) (err error) {
	for {
		if *isClose == true {
			return nil
		}
		buf := make([]byte, 1024, 1024)
		if n, err := from.Read(buf[:]); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		} else {
			if n > 0 {
				if _, err := to.Write(buf[:n]); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func Write(from io.Reader, to io.Writer, isClose *bool) (err error) {
	for {
		if *isClose == true {
			return nil
		}
		buf := make([]byte, 1024, 1024)
		if n, err := from.Read(buf[:]); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		} else {
			if n > 0 {
				if _, err := to.Write(buf[:n]); err != nil {
					return err
				}
			} else if n == 0 {
				continue
			}
		}
		return nil
	}
}
