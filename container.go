/*
   container.go - Building a container from scratch.
   Author: Hoanh An (hoanhan@bennington.edu)
   Date: 06/06/18

   Run commands:
      go run container.go run <cmd> <args>

      For example:
      go run container.go run /bin/bash

   Why would we want to build our own container?
      The point is not to build our own container engine that we are gonna use
      in production but to learn how the building blocks work.
      It's about understanding what namespaces are, what control group are, and
      how we can use chroot to look at the subset of directory systems from
      within the container.

   Steps:
      - Look at some characteristics of existing container and try to reproduce
      from within the container we write ourselves.
      - Start with the hostname namespace.
      - Change the filesytem that the container can see.
      - Look at namespacing and process IDs and how we need to interact with
      /proc directory to do that.
      - Use namespacing mounts for temporary filesystems.
      - Update control groups to limit what the container can use.
      - Make the container rootless.

   Setup:
      Need to use Ubuntu, instead of Mac is because some of the namespace files
      that we're gonna use are only defined for Go's Linux.

   Docker behavior:
      These are sample docker commands:
          docker run ubuntu:latest echo hello
          docker run -ti ubuntu:latest /bin/bash

      Comparing to docker commands, our program will be a lot similar:
          docker              run <image> <cmd> <args>
          go run container.go run         <cmd> <args>

   Container namespaces:
       Namespaces limit what we can see.
       Created with syscall.
       Inside the container, we can see a subset/some aspect of the whole machine.

   Chroot
      Chroot limits access to subset of directory tree on the host machine.
      We can setup a chroot to a directory on the host. From the container's
      point of view, that directory becomes its root directory.
*/

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// Check the first argument if it's run or child. Otherwise, just panic.
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
	// Print out the argument that are passing in for debugging.
	fmt.Printf("Running %v\n", os.Args[2:])

	// We need to change the hostname in the container before running the bash
	// profile to see changes.
	// However, by calling syscall.Sethostname([]byte("container")) in run, we
	// don't actually create a new namespace until we execute the cmd.Run
	// function. If we set the hostname before the Run function, we are setting
	// up the hostname for the host machine. Therefore, we need to create a
	// namespace first and then set the hostname.
	// A way to do that is to forking and execing twice. Create another
	// function called child. In run, we only create a namespace and in child,
	// we only set the hostname.
	// By doing exec.Command with /proc, we are calling the program again. Then
	// using append to pass in commands named child and all the arbitrary
	// arguments. In main function above, if we get child instead of run, we
	// can call the child function.
	// The first time we come in, we call run, then call the program again,
	// create a namespace. When we run it again, we will be given the child
	// parameter. So this time, we are not going to create any namepsaces but
	// set the hostname within the namespace we already created.
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// This is the process of namespacing the hostname. We can specify the
	// namespace we want by adding them to the cmd structure that we've setup.
	// Cloneflags are parameters that will be used on the clone syscall
	// function. Clone is actually what actually create a new process. Then we
	// are asking for NEWUTS namespace, where UTS stands for Unix Timestamp System.
	// We've built some element of containerization here. The container can
	// change its own hostname without affecting any other container or the host machine.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS,
	}

	// Not gonna execute until this Run function here.
	check(cmd.Run())
}

func child() {
	fmt.Printf("Running %v\n", os.Args[2:])

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Change the container hostname to container.
	check(syscall.Sethostname([]byte("container")))

	// We must create a sample filesystem beforehand in order for this to work.
	// Simply make a copy of the existing filesystem and name it sample_fs.
	// Then chroot points to that.
	check(syscall.Chroot("/home/sample_fs"))
	check(os.Chdir("/"))

	check(cmd.Run())
}

// check panics if anything go wrong.
func check(err error) {
	if err != nil {
		panic(err)
	}
}
