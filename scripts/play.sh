cd ~/chess/cutechess-cli-1.2.0-linux64/cutechess-cli
./cutechess-cli.sh -concurrency 6 \
  -pgnout ~/chess/counter41self.pgn \
  -engine conf=counter41 tc="inf" nodes=40000 \
  -engine conf=counter41 tc="inf" nodes=40000 \
  -ratinginterval 10 \
  -event SELF_PLAY_GAMES \
  -draw movenumber=40 movecount=10 score=20 \
  -resultformat per-color \
  -openings order=sequential start=1 format=epd file=~/chess/openings/openings_random.epd \
  -each proto=uci option.Hash=32 \
  -rounds 100000
