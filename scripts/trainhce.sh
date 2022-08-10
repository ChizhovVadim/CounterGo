#!/bin/bash

go run ./cmd/trainhce -eval linear \
    -td ~/chess/fengen.txt \
    -vd ~/chess/tuner/quiet-labeled.epd \
    -threads 4 \
    -epochs 100 \
    -sw 0.5 \
    -dms 8000000
