package main

/*
#include <stdio.h>
static void p(char *s) {
printf("buff is [%s]\n", s);
}

static void pl(char *s, int n) {
for(int i = 0; i <n ;i++ ) {
printf("%d ", s[i]);
}

printf("\n");
}
*/
import "C"

import (
    "fmt"
    "github.com/shanicky/memkind-go/memkind"
    "log"
    "math/rand"
    "os"
    "reflect"
    "sync/atomic"
    "time"
    "unsafe"
)

var (
    alloc    = int64(0)
    parallel = 64
)

func main() {
    path := "/tmp"

    if len(os.Args) > 2 {
	fmt.Printf("Usage: %s [pmem_kind_dir_path]", os.Args[0])
	os.Exit(1)
    }

    if len(os.Args) == 2 {
	path = os.Args[1]
    }

    fmt.Printf("PMEM kind directory: %s\n", path);
    var kind memkind.MemkindT = nil
    if errno := memkind.MemkindCreatePmem(path, 1024*1024*1024*32, &kind); errno != 0 {
	memkindPrintError(errno)
	os.Exit(1)
    }

    defer memkind.MemkindDestroyKind(kind)

    buff := make([]byte, 1024*1024*100)

    for i := 0; i < len(buff); i++ {
	buff[i] = 'x'
    }

    go func() {
	ticker := time.NewTicker(time.Second)

	last := alloc
	for range ticker.C {
	    current := atomic.LoadInt64(&alloc)
	    fmt.Printf("%.2fMB/s\n", float64(current-last)/float64(1024*1024))
	    last = current
	}
    }()

    for i := 0; i < parallel; i++ {
	go func() {
	    for {
		size := uint(rand.Int() % (1024 * 1024 * 10))
		ptr := memkind.MemkindMalloc(kind, size)
		if ptr == nil {
		    log.Fatal("ptr is nil")
		}

		bs := unsafePtrToBytes(ptr, size)
		copy(bs, buff[:size])
		memkind.MemkindFree(kind, ptr)


		atomic.AddInt64(&alloc, int64(size))
	    }

	}()
    }
    select {}

    // printPtr(ptr)
}

func printPtr(ptr unsafe.Pointer) {
    C.p((*C.char)(ptr))
    C.pl((*C.char)(ptr), C.int(10))
}

func unsafePtrToBytes(ptr unsafe.Pointer, size uint) []byte {
    return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
	Data: uintptr(ptr), Len: int(size), Cap: int(size),
    }))
}

func memkindPrintError(errno int32) {
    buf := make([]byte, 1024)
    memkind.MemkindErrorMessage(errno, buf, 1024)
    fmt.Printf("%s\n", buf)
}
