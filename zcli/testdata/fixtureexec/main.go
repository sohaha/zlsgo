package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "stream":
			fmt.Println("fixture-stream")
			fmt.Fprintln(os.Stderr, "fixture-err")
			return
		case "cat":
			// Mirrors the value passed on stdin so KeepRun input can be tested.
			_, _ = io.Copy(os.Stdout, os.Stdin)
			return
		case "readline":
			line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			fmt.Print(line)
			return
		case "sleep":
			time.Sleep(30 * time.Second)
			return
		case "orphan":
			// Spawn a grandchild that outlives this process while holding the
			// inherited stdout, to exercise WaitDelay pipe draining.
			grandchild := exec.Command(os.Args[0], "sleep")
			grandchild.Stdout = os.Stdout
			grandchild.Stderr = os.Stderr
			if err := grandchild.Start(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			fmt.Println("orphan-started")
			return
		}
	}

	fmt.Printf("%s|%s", os.Getenv("EXEC_TEST_TOKEN"), mustWd())
	for _, arg := range os.Args[1:] {
		fmt.Printf("|%s", arg)
	}
	fmt.Println()
}

func mustWd() string {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return wd
}
