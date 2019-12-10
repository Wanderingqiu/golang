#!/bin/bash

go run client.go -f 1.txt -print y&
go run client.go -f 2.txt&
go run client.go -f 3.txt&
go run client.go -f 4.txt
