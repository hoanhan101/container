package main

import (
	"filepath"
	"fmt"
	"ioutil"
	"os"
	"os/exec"
	"strconv"
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
	// - CLONE_NEWPID: Process IDs namespace.
	// - CLONE_NEWNS: Mount namespace to make the mount point only visible
	//   inside the container.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}

	// Not gonna execute until this Run function here.
	check(cmd.Run())
}

func child() {
	fmt.Printf("Running %v\n", os.Args[2:])

	// Set control group.
	set_cgroup()

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
	// `/proc` isn't just a regular file system but a pseudo file system. It does not
	// contain real files but rather runtime system information (e.g. system
	// memory, devices mounted, hardware configuration, etc). It is a mechanism the
	// kernel uses to communicate information about running processes from the kernel
	// space into user spaces. From the user space, it looks like a normal file
	// system.
	//
	// Because we change the root inside the container, we don't currently have this
	// `/procs` file system available. Therefore, we need to mount it.
	check(syscall.Mount("proc", "proc", "proc", 0, ""))

	// Mount a temporary file system (tmpfs). It looks and behaves just like a normal
	// file system. However, rather writing files and directories to disk, it
	// holds in memory.
	//
	// Name the temporary file system temp_fs, mount it to our existing my_temp
	// directory, specify that we want a tmpfs.
	check(syscall.Mount("temp_fs", "my_temp", "tmpfs", 0, ""))

	check(cmd.Run())

	// Clean up after run.
	check(syscall.Unmount("proc", 0))
	check(syscall.Unmount("my_temp", 0))
}

// If namespace limits what users can see then control group limits what users can
// use. For example, the mount of CPU, memory or network IO.
//
// Control groups are mounted as pseudo file system. To look at the control groups
// that are mounted on the host machine, execute `mount | grep cgroup`
//
// Control groups are hierarchical. For example, there exists a directory called
// docker inside `/sys/fs/cgroup/memory`. If we look inside it, we will see a
// whole other set of premises, applying to the members of docker cgroup.
//
// By default, processes get assigned to the top level control group. For example,
// For memory, each process get written to cgroup.procs by default. If we want to
// assign a process to a particular control group, we can write it into the sub
// directory for that control group, within that cgroup.procs file. All the
// children of that process will also be assigned to the same process. Processes
// assigned to that directory within control group will inherit settings from
// their parents.
func set_cgroup() {
	// Control group mounted on our host machine.
	cgroups := "/sys/fs/cgroup/"

	// Write to memory control group hierarchy.
	mem := filepath.Join(cgroups, "memory")

	// Create (if doesn't exist) a sub directory named hoanh
	os.Mkdir(filepath.Join(mem, "hoanh"), 0755)

	// Inside that, write to 3 files: memory limit_in_bytes, swap memory
	// limit_in_bytes, notify_on_release (if there is no more processes in left
	// inside the control group, can delete the control group).
	// 999424 is basically 1 MB.
	check(ioutil.WriteFile(filepath.Join(mem, "hoanh/memory.limit_in_bytes"),
		[]byte("999424"), 0700))
	check(ioutil.WriteFile(filepath.Join(mem, "hoanh/memory.memsw.limit_in_bytes"),
		[]byte("999424"), 0700))
	check(ioutil.WriteFile(filepath.Join(mem, "hoanh/notify_on_release"),
		[]byte("1"), 0700))

	// Get the current process id and write it to the cgroup.procs file.
	pid := strconv.Itoa(os.Getpid())
	check(ioutil.WriteFile(filepath.Join(mem, "hoanh/cgroup.procs"),
		[]byte(pid), 0700))
}

// check panics if anything go wrong.
func check(err error) {
	if err != nil {
		panic(err)
	}
}
