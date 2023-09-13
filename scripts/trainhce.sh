#!/bin/bash

go run ./cmd/trainhce -eval linear \
    -td ~/chess/fengen-counter41.txt \
    -vd ~/chess/tuner/quiet-labeled.epd \
    -epochs 200 \
    -dms 8000000
