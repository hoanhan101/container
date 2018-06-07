# container

Sample code on how to build a container from scratch using Go.

**Reference:**

- [Building Containers from Scratch with Go by Liz
  Rice](https://www.safaribooksonline.com/library/view/building-containers-from/9781491988404/)

**Run:**

```
go run container.go run <cmd> <args>
```

For example:

```
go run container.go run /bin/bash
```

**Why would we want to build our own container?**

The point is not to build our own container engine that we are gonna use
in production but to learn how the building blocks work.
It's about understanding what namespaces are, what control group are, and
how we can use chroot to look at the subset of directory systems from
within the container.

**Steps:**

- Look at some characteristics of existing container and try to reproduce
from within the container we write ourselves.
- Start with the hostname namespace.
- Change the file system that the container can see.
- Look at namespacing and process IDs and how we need to interact with
/proc directory to do that.
- Use namespacing mounts for temporary file system.
- Update control groups to limit what the container can use.
- Make the container rootless.

**Setup:**

Need to use Ubuntu, instead of Mac is because some of the namespace files
that we're gonna use are only defined for Go's Linux.

**Docker behavior:**

These are some sample docker commands:

```
docker run ubuntu:latest echo hello
docker run -ti ubuntu:latest /bin/bash
```

Comparing to docker commands, our program will be a lot similar:

```
docker              run <image> <cmd> <args>
go run container.go run         <cmd> <args>
```
