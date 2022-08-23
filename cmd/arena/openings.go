package main

var openings = []string{
	"rn1q1rk1/1p2ppbp/p1p2np1/3p4/2PP2b1/1PNBPN2/P4PPP/R1BQ1RK1 w - - 1 9 ",
	"r1b1kb1r/1pq2ppp/p1nppn2/8/3NP1P1/2N4P/PPP2PB1/R1BQK2R w KQkq - 2 9 ",
	"r1bq1rk1/pp1nppbp/3p1np1/8/P2p1B2/4PN1P/1PP1BPP1/RN1Q1RK1 w - - 0 9 ",
	"r1bqk2r/p3bpp1/1pn1pn1p/2pp4/3P3B/2PBPN2/PP1N1PPP/R2QK2R w KQkq - 0 9 ",
	"r2qk2r/p1pp1ppp/b1p2n2/8/2P5/6P1/PP1QPP1P/RN2KB1R w KQkq - 1 9 ",
	"r1bqr1k1/pppp1ppp/2n2n2/2bN4/2P1p2N/6P1/PP1PPPBP/R1BQ1RK1 w - - 6 9 ",
	"r1bq1rk1/1p3ppp/2n1pn2/p1bp4/2P5/P3PN2/1P1NBPPP/R1BQK2R w KQ - 2 9 ",
	"r1bq1rk1/ppp1p1bp/3p1np1/n2P1p2/2P5/2N2NP1/PP2PPBP/R1BQ1RK1 w - - 1 9 ",
	"r1bq1rk1/ppp2ppp/2np1n2/4p3/2P5/2PP1NP1/P3PPBP/R1BQ1RK1 w - - 1 9",
	"r1bqk1nr/1pp2pbp/p2p2p1/1N1P4/2PpP3/8/PP3PPP/R1BQKB1R w KQkq - 0 9 ",
	"rn1qkb1r/pbpp1p2/1p2p2p/6p1/2PP4/2N1PNn1/PP3PPP/R2QKB1R w KQkq - 0 9 ",
	"rn1qk2r/pp2bppp/2p2nb1/3p4/3N4/3P2P1/PP2PPBP/RNBQ1RK1 w kq - 3 9 ",
	"rn1q1rk1/pbp1bppp/1p1pp3/7n/2PP4/2N1PNB1/PPQ2PPP/R3KB1R w KQ - 0 9 ",
	"r1bq1rk1/bpp2ppp/p1np1n2/4p3/B3P3/2PP1N2/PP1N1PPP/R1BQ1RK1 w - - 2 9 ",
	"r3kbnr/pp1b1ppp/1q2p3/3pP3/3n4/2P2N2/PP3PPP/R1BQKB1R w KQkq - 0 9 ",
}

var opening2 = []string{
	//Французская
	"1. e4 e6 2. d4 d5 3. Nc3 Nf6 4. e5 Nfd7 5. f4 c5 6. Nf3 Nc6 7. Be3",
	"1. e4 e6 2. d4 d5 3. Nc3 Bb4 4. e5 c5 5. a3 Bxc3+ 6. bxc3",
	"1. e4 e6 2. d4 d5 3. Nc3 dxe4 4. Nxe4",
	"1. e4 e6 2. d4 d5 3. e5 c5 4. c3 Nc6 5. Nf3 Qb6 6. Bd3",
	//Сицилианская
	"1. e4 c5 2. Nf3 d6 3. d4 cxd4 4. Nxd4 Nf6 5. Nc3 Nc6",
	"1. e4 c5 2. Nf3 d6 3. d4 cxd4 4. Nxd4 Nf6 5. Nc3 a6",
	//Открытые
	"1. e4 e5 2. Nf3 Nf6 3. Nxe5",
	"1. e4 e5 2. Nf3 Nf6 3. d4",
	"1. e4 e5 2. Nf3 Nc6 3. d4 exd4 4. Nxd4",
	"1. e4 e5 2. Nf3 Nc6 3. Bc4 Bc5",
	"1. e4 e5 2. Nf3 Nc6 3. Bb5",
	//Филидор
	"1. e4 d6 2. d4 Nf6 3. Nc3 e5 4. dxe5 dxe5 5. Qxd8+ Kxd8",
	//Скандинавка
	"1. e4 d5",
	//Каро-Канн
	"1. e4 c6 2. Nf3 d5 3. d3 dxe4 4. dxe4 Qxd1+ 5. Kxd1",
	"1.e4 c6 2.d4 d5 3.e5 c5",
	"1.e4 c6 2.d4 d5 3.e5 Bf5",
	"1.e4 c6 2.d4 d5 3.exd5 cxd5",
	"1.e4 c6 2.d4 d5 3.Nc3 dxe4 4.Nxe4 Nf6",
	"1.e4 c6 2.d4 d5 3.Nc3 dxe4 4.Nxe4 Bf5",
	"1.e4 c6 2.d4 d5 3.Nc3 dxe4 4.Nxe4 Nd7",
	//Пирц
	"1. e4 d6 2. d4 Nf6 3. Bd3",
	"1. e4 d6 2. d4 Nf6 3. Nc3 g6 4. Bg5",
	"1. e4 d6 2. d4 Nf6 3. Nc3 g6 4. Be3 c6 5. Qd2",
	"1. e4 d6 2. d4 Nf6 3. Nc3 g6 4. f4 Bg7 5. Nf3 c5",
	"1. e4 d6 2. d4 Nf6 3. Nc3 g6 4. Nf3",
	//Нимцович
	"1.d4 Nf6 2.c4 e6 3.Nc3 Bb4",
	//Грюнфельд
	"1. d4 Nf6 2. c4 g6 3. Nc3 d5 4. cxd5 Nxd5 5. e4 Nxc3 6. bxc3 Bg7",
	//Эндшпиль 1. d4 d6
	"1. d4 d6 2. c4 e5 3. dxe5 dxe5 4. Qxd8+ Kxd8",
	//Славянская
	"1. d4 d5 2. c4 c6 3. Nf3 Nf6 4. Nc3 dxc4 5. a4 Bf5 6. Ne5",
	"1. d4 d5 2. c4 c6 3. Nf3 Nf6 4. Nc3 dxc4 5. a4 Bf5 6. e3 e6 7. Bxc4 Bb4 8. O-O",
	"1. d4 d5 2. c4 c6 3. Nf3 Nf6 4. Nc3 e6 5. e3 Nbd7 6. Qc2",
	"1. d4 d5 2. c4 c6 3. Nf3 Nf6 4. Nc3 e6 5. e3 Nbd7 6. Bd3 dxc4 7. Bxc4 b5",
	//Принятый ферзевый гамбит
	"1. d4 d5 2. c4 dxc4",
	//Волжский гамбит
	"1. d4 Nf6 2. c4 c5 3. d5 b5 4. cxb5 a6 5. bxa6 g6 6. Nc3 Bg7 7. e4 O-O 8. a7 Rxa7",
	"1. d4 Nf6 2. c4 c5 3. d5 b5 4. cxb5 a6 5. b6",
	//Будапешт
	"1. d4 Nf6 2. c4 e5 3. dxe5 Ng4",
}
