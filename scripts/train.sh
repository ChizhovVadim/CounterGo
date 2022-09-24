#!/bin/bash

go run ./cmd/train \
    -td ~/chess/fengen-counter41.txt \
    -vd ~/chess/tuner/quiet-labeled.epd \
    -net ~/chess/net \
    -threads 6 \
    -epochs 30 \
    -sw 0.75 \
    -dms 15000000
