package main

import (
	"bufio"
	"os"
)

/*
#include "sys/mman"
int create_mem_fd() {
       return memfd_create("foo", 0);
}

void
 */
import "C"

func createMemFd() {
	syscall.Create
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for sd
}
