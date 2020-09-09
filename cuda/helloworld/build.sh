#!/bin/bash

nvcc \
  --ptxas-options=-v \
  --compiler-options '-fPIC' -o libmaxmul.so --shared maxmul.cu
