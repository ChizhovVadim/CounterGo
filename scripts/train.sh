#!/bin/bash

go run ./cmd/train \
    -td ~/chess/fengen-counter41.txt \
    -vd ~/chess/tuner/quiet-labeled.epd \
    -net ~/chess/net \
    -epochs 30 \
    -dms 50000000
