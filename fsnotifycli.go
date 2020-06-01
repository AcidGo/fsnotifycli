package main

import (
    "flag"
    "fmt"
    "log"
    "io"
    "os"

    "github.com/fsnotify/fsnotify"
)

const (
    writeOrCreateMask = fsnotify.Write | fsnotify.Create
    writeMask = fsnotify.Write
    MaxContentSize = 1024*10
    PerFileBufferSize = 100
)

var (
    path string
    hasContent bool
    hasHelp bool
    contentSize int
)

func usage() {
    fmt.Fprintf(os.Stderr, `fsnotifycli version: 0.0.1
Author: AcidGo
Usage: fsnotifycli [-hc] [-s <contentsize>] <path>

Options:
`)
    flag.PrintDefaults()
}

func parseFlag() {
    flag.BoolVar(&hasContent, "c", false, "show the created file content")
    flag.BoolVar(&hasHelp, "h", false, "show the help message")
    flag.IntVar(&contentSize, "s", 1024, "set read `contentsize` in B")

    flag.Usage = usage

    flag.Parse()

    if hasHelp {
        flag.Usage()
        os.Exit(0)
    }

    if flag.NArg() > 1 {
        flag.Usage()
        os.Exit(1)
    }

    path = flag.Arg(0)
    if path == "" {
        path = "./"
    }

    if contentSize > MaxContentSize {
        log.Printf("the MaxContentSize is %d, so the contentSize becomes %d\n", MaxContentSize, MaxContentSize)
        contentSize = MaxContentSize
    }

    log.Printf("The notify path is: %s\n", path)
    log.Println("enjoy it~~~")
}

func readContent(filePath string) {
    file, err := os.Open(filePath)
    if err != nil {
        log.Println("error when open file:", err)
        return
    }
    defer file.Close()

    var bufferSize int
    NcontentSize := contentSize
    if contentSize < PerFileBufferSize {
        bufferSize = contentSize
    } else {
        bufferSize = PerFileBufferSize
    }
    buffer := make([]byte, bufferSize)

    log.Printf("content of %s:\n", filePath)
    for {
        bytesread, err := file.Read(buffer)
        if err != nil {
            if err != io.EOF {
                log.Println("error when read file:", err)
            } else {
                fmt.Println()
                log.Printf("end of read file\n")
            }
            break
        }
        fmt.Print(string(buffer[:bytesread]))
        NcontentSize -= bytesread
        if NcontentSize <= 0 {
            fmt.Println()
            fmt.Println("...")
            log.Printf("end of read file\n")
            break
        }
    }
}

func init() {
    log.SetFlags(log.Ldate|log.Lmicroseconds)
}

func main() {
    parseFlag()

    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    defer watcher.Close()

    done := make(chan bool)
    go func() {
        for {
            select {
            case event, ok := <-watcher.Events:
                if ! ok {
                    return
                }
                log.Println("event:", event)
                if hasContent && (event.Op & writeMask != 0) {
                    go readContent(event.Name)
                }
            case err, ok := <-watcher.Errors:
                if ! ok {
                    return
                }
                log.Println("error:", err)
            }
        }
    }()

    err = watcher.Add(path)
    if err != nil {
        log.Fatal(err)
    }
    <-done
}

