go build ./cmd/tests
./tests -name quality \
    -eval nnue \
    -testpath ~/chess/tests/tests.epd \
    -vd ~/chess/tuner/quiet-labeled.epd
