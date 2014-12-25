#! /bin/sh
rm -rf *~
go build || exit
go install || exit

