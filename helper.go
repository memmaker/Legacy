package main

import (
    "bufio"
    "image"
    "io"
    "io/fs"
    "log"
)

func mustLoadImage(filename string) image.Image {
    img, _, err := image.Decode(mustOpen(filename))
    if err != nil {
        log.Fatal(err)
    }
    return img
}
func doesFileExist(filename string) bool {
    _, err := fs.Stat(assetFS, filename)
    return err == nil
}
func mustOpen(filename string) io.ReadCloser {
    f, err := assetFS.Open(filename)
    if err != nil {
        log.Fatal(err)
    }
    return f
}

func readLines(filename string) []string {
    file, err := assetFS.Open(filename)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    var lines []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }
    if err := scanner.Err(); err != nil {
        log.Fatal(err)
    }
    return lines
}
