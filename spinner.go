package main

import (
	"fmt"
	"github.com/schollz/progressbar"
)

func main() {
	count_all:=10000000000
	count:=10000000000
	percent :=count_all/100
	bar := progressbar.New(100)
	for  ; count> 0; count-- {
		if(count%percent==0){
			bar.Add(1)
		}
		
	}
	fmt.Println()
}
