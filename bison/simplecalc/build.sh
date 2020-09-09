#/bin/bash

# cc -Wall -fPIC --shared -lm -o libcalc.so calc.tab.c

bison calc.y
go build
