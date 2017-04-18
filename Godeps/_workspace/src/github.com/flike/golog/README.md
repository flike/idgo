# golog

simple log library for Golang

## case 

```
package main

import (
	"fmt"
	. "github.com/flike/golog"
	"os"
)

func main() {
	path := "/tmp/test_log"
	os.RemoveAll(path)
	os.Mkdir(path, 0777)
	fileName := path + "/test.log"

	h, err := NewRotatingFileHandler(fileName, 1024*1024, 2)
	if err != nil {
		fmt.Println(err.Error())
	}
	//GlobalLogger is a global variable
	GlobalLogger = New(h, Lfile|Ltime|Llevel)
	GlobalLogger.SetLevel(LevelTrace)
	args1 := "go"
	args2 := "log"
	args3 := "golog"
	//the log will record into the file(/tmp/test_log/test.log)
	Debug("Mode", "main", "OK", 0, "args1", args1, "args2", args2, "args3", args3)
	GlobalLogger.Close()
}

```
