#!/bin/bash

go run ./cmd/trainhce -eval linear \
    -td ~/chess/fengen-counter41.txt \
    -vd ~/chess/tuner/quiet-labeled.epd \
    -threads 6 \
    -epochs 200 \
    -sw 0.75 \
    -dms 8000000
