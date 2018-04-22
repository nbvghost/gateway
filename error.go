package main

import (
	"log"
	"runtime"
	"strconv"
)

func CheckError(err error) string {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		log.Println(file, line, err)
		return file + " " + strconv.Itoa(line) + " " + err.Error()
	} else {
		return ""
	}
}
func Trace(v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	//logic.Trace(funcName,file,line,ok)
	log.Println(file, line, v)

}
