![logo](https://raw.githubusercontent.com/ChizhovVadim/CounterGo/master/logo.png)
# Counter
Counter is a chess engine.
+ free
+ open source
+ support for different OS (Linux, macOS, Windows)
+ multi threading support
+ [UCI](http://www.shredderchess.com/chess-info/features/uci-universal-chess-interface.html) protocol support. You can use any [chess GUI interface](https://www.chessprogramming.org/UCI#GUIs) that supports UCI protocol

## Strength

Chess Rating lists:
+ [CCRL](https://ccrl.chessdom.com/ccrl/)
  + [CCRL 40/15 progress](http://www.computerchess.org.uk/ccrl/4040/cgi/compare_engines.cgi?family=Counter&print=Rating+list&print=Results+table&print=LOS+table&print=Ponder+hit+table&print=Eval+difference+table&print=Comopp+gamenum+table&print=Overlap+table&print=Score+with+common+opponents)
  + [CCRL 40/2 progress](http://www.computerchess.org.uk/ccrl/404/cgi/compare_engines.cgi?family=Counter&print=Rating+list&print=Results+table&print=LOS+table&print=Ponder+hit+table&print=Eval+difference+table&print=Comopp+gamenum+table&print=Overlap+table&print=Score+with+common+opponents)
+ [FGRL](http://fastgm.de/)
+ [CEGT](http://www.cegt.net/)
+ [Gambit Rating List](http://rebel13.nl/grl-best-40-2.html)

|Version|GRL  |CCRL 40/15|FastGM 60+0.6|CEGT 40/4|
|-------|-----|----------|-------------|---------|
|3.9    |3065 |          |             |2959     |
|3.8    |2994 |3012      |2817         |2887     |
|3.7    |2972 |2970      |2784         |2854     |
|3.6    |     |          |2757         |2820     |
|3.5    |     |2907      |2718         |2777     |
|3.4    |     |2881      |2679         |2742     |
|3.3    |     |2847      |2647         |2700     |
|3.2    |     |2834      |2624         |2692     |

## TCEC
Counter is participating in [TCEC](https://wiki.chessdom.org/Main_Page) tournament.
+ [Season 20](https://wiki.chessdom.org/TCEC_Season_20_Engines)
+ [Season 19](https://wiki.chessdom.org/TCEC_Season_19_Engines)
+ [Season 18](https://wiki.chessdom.org/TCEC_Season_18_Engines)
+ [Season 17](https://wiki.chessdom.org/TCEC_Season_17_Engines)

## Technical description
Currently Counter is an alpha beta engine with Hand Crafted Eval. Counter uses bitboards for board representation and Magic bitboards for move generation. Multithreading implemented with LazySMP method. Counter written in the [go](https://golang.org/) programming language. I think programming should be fun. And C/C++ is not funny at all.

## How to write chess engine
### Level0
- able to generate legal moves and select random move
- implement uci protocol
- unit test for move generator
### Level1 (only exact search methods)
- PESTO eval
- Iterative deepening
- alphabeta, QS(good captures and check escapes)
- Transposition table
- Internal iterative deepening
- move order: trans move, good captures, killers, bad captures and history
- repeat detect
- Aspiration window
- PVS in root
- simple time manager
- mate distance pruning
### Level2 (simple methods with maximum ELO increase)
- NMP R=4+d/6, null move case in repeat detect
- LMR R~log(d)log(m). In ideal case without lmr research, search tree will growth linear
- Leaf prunings (reverse futility pruning, Late move pruning, SEE pruning)
- Singular extension, check extension
- eval (Material, King safety, passed pawns, threats, PSQT, mobility). Texel tuning
- LazySMP
### Level3 (complex methods or methods with low ELO increase)
- complicated time manager
- NNUE eval
- endgame TB
- PVS
- performance (increment staged move generator, store static eval in TT or cache static eval, pawn hash table in eval, other optimizations)
- Probcut
- ...

---------------------------------------------------------------

Counter Copyright (c) Vadim Chizhov. All rights reserved.
