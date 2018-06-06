/*
   container.go - Building a container from scratch.
   Reference: O'Reilly Safari's Building Containers from Scratch with Go by Liz Rice
   Author: Hoanh An (hoanhan@bennington.edu)
   Date: 06/06/18

   Sample run:
       go run container.go run <cmd> <args>
       go run container.go run /bin/bash
*/

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("Exit.")
	}
}

func run() {
	fmt.Printf("Running %v\n", os.Args[2:])

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS, // Unix Timestamp System
	}

	must(cmd.Run())
}

func child() {
	fmt.Printf("Running %v\n", os.Args[2:])

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	must(syscall.Sethostname([]byte("container")))

	must(cmd.Run())
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
