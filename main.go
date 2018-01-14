package main

import (
	"fmt"
	"log"
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
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}
}

func child() {
	fmt.Printf("start child......, pid %v\n", syscall.Getpid())
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

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
		log.Printf("failed to mount proc to %s: %v", target, err)
		panic(err)
	}
}
