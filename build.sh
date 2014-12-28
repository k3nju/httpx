#! /bin/sh
rm -rf *~
go test -v || exit
go build || exit
go install || exit

