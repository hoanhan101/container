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

	// Namespaces limit what we can see. Created with syscall.
	// Inside the container, we can see a subset/some aspect of the whole machine.
	//
	// This is the process of namespacing the hostname. We can specify the
	// namespace we want by adding them to the cmd structure that we've setup.
	// Cloneflags are parameters that will be used on the clone syscall
	// function. Clone is actually what actually create a new process.
    // - CLONE_NEWUTS: UTS namespace, where UTS stands for Unix Timestamp System.
    // - CLONE_NEWPID: Process IDs namespace
	cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID,
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

	// Chroot limits access to subset of directory tree on the host machine.
	// We can setup a chroot to a directory on the host. From the container's
	// point of view, that directory becomes its root directory.
	//
	// We must create a sample filesystem beforehand in order for this to work.
	// Simply make a copy of the existing filesystem and name it sample_fs.
	// Then chroot points to that.
	check(syscall.Chroot("/home/sample_fs"))
	check(os.Chdir("/"))

	// When we execute `ps`, we are able to see the process IDs. It will look inside
	// the directory called `/proc` to get process informations. From the host machine
	// point of view, we can see all the processes running on it.
	//
	// `/proc` isn't just a regular file system but a pseudo-file system. It does not
	// contain real files but rather runtime system information (e.g. system
	// memory, devices mounted, hardware configuration, etc). It is a mechanism the
	// kernel uses to communicate information about running processes from the kernel
	// space into user spaces. From the user space, it looks like a normal file
	// system.
	//
	// Because we change the root inside the container, we don't currently have this
	// `/procs` file system available. Therefore, we need to mount it.
	check(syscall.Mount("proc", "proc", "proc", 0, ""))

	check(cmd.Run())

	// Clean up after run.
	check(syscall.Unmount("proc", 0))
}

// check panics if anything go wrong.
func check(err error) {
	if err != nil {
		panic(err)
	}
}
