package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"syscall"
)

func main() {
	switch os.Args[1] {
	case "run":
		parent()
	case "child":
		child()
	default:
		panic("what should I do")
	}
}

func parent() {
	cmd := exec.Command(os.Args[0], append([]string{"child"}, os.Args[2:]...)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		panic(fmt.Sprintf("start parent err: %v", err))
	}

	log.Printf("container PID: %d", cmd.Process.Pid)
	// set bridge and veth pair for container.
	if err := putIface(cmd.Process.Pid); err != nil {
		panic(fmt.Sprintf("putIface err: %v", err))
	}

	if err := cmd.Wait(); err != nil {
		panic(fmt.Sprintf("wait parent err: %v\n", err))
	}
}

func child() {
	fmt.Printf("start child......, pid %v\n", syscall.Getpid())
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	// setup environment for container.
	setup()
	if err := cmd.Run(); err != nil {
		panic(fmt.Sprintf("child panic: %v", err))
	}
}

func setup() {
	if err := syscall.Sethostname([]byte("container")); err != nil {
		panic(fmt.Sprintf("Sethostname: %v", err))
	}

	pwd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("get pwd err: %v\n", err))
	}

	target := path.Join(pwd, "rootfs")
	if err := syscall.Chroot(target); err != nil {
		panic(fmt.Sprintf("chroot err: %v\n", err))
	}
	if err := os.Chdir("/"); err != nil {
		panic(fmt.Sprintf("chdir err: %v\n", err))
	}

	if err := syscall.Mount("proc", "proc", "proc", 0, ""); err != nil {
		panic(fmt.Sprintf("failed to mount proc to %s: %v", target, err))
	}

	lnk, err := waitForIface()
	if err != nil {
		panic(fmt.Sprintf("waitForIface err: %v", err))
	}

	if err := setupIface(lnk, fmt.Sprintf(ipTmpl, rand.Intn(253)+2)); err != nil {
		panic(fmt.Sprintf("setupIface err: %v", err))
	}
}
