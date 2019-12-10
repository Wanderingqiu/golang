package main

import(
    "os"
)

func main(){
    str := "\n456"
    filename := "1.txt"
    fd,_:=os.OpenFile(filename,os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
    _,_ = fd.WriteString(str)
    fd.Close()
}
