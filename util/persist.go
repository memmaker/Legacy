package util

import (
    "os"
    "path"
)

func Persist(key, value string) {
    filename := path.Join("persist", key)
    err := os.WriteFile(filename, []byte(value), 0666)
    if err != nil {
        println("Error writing file", filename, err.Error())
    }
}

func Get(key string) string {
    fileData, err := os.ReadFile(path.Join("persist", key))
    if err != nil {
        return ""
    }
    return string(fileData)
}
