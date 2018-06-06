/*
    Build a container from scratch.
    Run:
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
    default:
        panic("Exit.")
    }
}

func run() {
    fmt.Printf("Running %v\n", os.Args[2:])

    cmd := exec.Command(os.Args[2], os.Args[3:]...)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.SysProcAttr = &syscall.SysProcAttr {
        Cloneflags: syscall.CLONE_NEWUTS, // Unix Timestamp System
    }

    must(cmd.Run())
}

func must(err error) {
    if err != nil {
        panic(err)
    }
}
