![logo](https://raw.githubusercontent.com/ChizhovVadim/CounterGo/master/logo.png)
# Counter
Counter is a free, open-source chess engine, implemented in [Go](https://golang.org/).
Counter supports standard UCI (universal chess interface) protocol.

## Strength

Chess Rating lists:
+ [CCRL](https://ccrl.chessdom.com/ccrl/)
  + [CCRL 40/2](http://www.computerchess.org.uk/ccrl/404/cgi/compare_engines.cgi?family=Counter&print=Rating+list&print=Results+table&print=LOS+table&print=Ponder+hit+table&print=Eval+difference+table&print=Comopp+gamenum+table&print=Overlap+table&print=Score+with+common+opponents)
  + [CCRL 40/15](http://www.computerchess.org.uk/ccrl/4040/cgi/compare_engines.cgi?family=Counter&print=Rating+list&print=Results+table&print=LOS+table&print=Ponder+hit+table&print=Eval+difference+table&print=Comopp+gamenum+table&print=Overlap+table&print=Score+with+common+opponents)
+ [FGRL](http://fastgm.de/)
+ [CEGT](http://www.cegt.net/)
+ [Gambit Rating List](http://rebel13.nl/grl-best-40-2.html)

|Version|GRL  |CCRL 40/15|FastGM 60+0.6|CEGT 4/40|
|-------|-----|----------|-------------|---------|
|3.9    |3065 |          |             |         |
|3.8    |2994 |3012      |2817         |2887     |
|3.7    |2972 |2970      |2784         |2854     |
|3.6    |     |          |2757         |2820     |
|3.5    |     |2907      |2718         |2777     |
|3.4    |     |2881      |2679         |2742     |
|3.3    |     |2847      |2647         |2700     |
|3.2    |     |2834      |2624         |2692     |

## Commands
Counter supports [UCI protocol](http://www.shredderchess.com/chess-info/features/uci-universal-chess-interface.html) commands.

## Features
### Board
+ Magic bitboards
### Evaluation
+ Texel's Tuning Method
### Search
+ Parallel search (Lazy SMP)
+ Iterative Deepening
+ Aspiration Windows
+ Transposition Table
+ Null Move Pruning
+ Late Move Reductions
+ Futility Pruning
+ Move Count Based Pruning
+ SEE Pruning
+ Singular extensions

## Information about chess programming
+ [Chess Programming Wiki](https://www.chessprogramming.org)
+ [Bruce Moreland's Programming Topics](https://web.archive.org/web/20071026090003/http://www.brucemo.com/compchess/programming/index.htm)

## Thanks
+ Vladimir Medvedev, GreKo
+ Fabien Letouzey, Fruit and Senpai
+ Robert Hyatt, Crafty

---------------------------------------------------------------

Counter Copyright (c) Vadim Chizhov. All rights reserved.
