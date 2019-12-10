#!/bin/bash

go run worker.go -i 172.17.0.15:9997 -e 122.51.83.192:9997&
go run worker.go -i 172.17.0.15:9998 -e 122.51.83.192:9998&
go run worker.go -i 172.17.0.15:9999 -e 122.51.83.192:9999

