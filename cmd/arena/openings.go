package main

import "strings"

const openings = `
//Французская
1. e4 e6 2. Qe2
1. e4 e6 2. d3 d5 3. Nd2 Nf6 4. Ngf3
1. e4 e6 2. Nf3 d5 3. Nc3
1. e4 e6 2. d4 d5 3. exd5 exd5
1. e4 e6 2. d4 d5 3. Bd3
1. e4 e6 2. d4 d5 3. Nd2
1. e4 e6 2. d4 d5 3. Nc3 Nf6 4. e5 Nfd7 5. Nce2
1. e4 e6 2. d4 d5 3. Nc3 Nf6 4. e5 Nfd7 5. f4 c5 6. Nf3 Nc6 7. Be3
1. e4 e6 2. d4 d5 3. Nc3 Bb4 4. e5 c5 5. a3 Bxc3+ 6. bxc3
1. e4 e6 2. d4 d5 3. Nc3 dxe4 4. Nxe4
1. e4 e6 2. d4 d5 3. e5 c5 4. c3 Nc6 5. Nf3 Qb6 6. Bd3
1. e4 e6 2. d4 d5 3. e5 c5 4. c3 Nc6 5. Nf3 Qb6 6. a3
1. e4 e6 2. d4 d5 3. e5 c5 4. c3 Nc6 5. Nf3 Qb6 6. Be2
//Сицилианская
1. e4 c5 2. c3
1. e4 c5 2. Nf3 Nc6 3. Bb5
1. e4 c5 2. Nf3 Nc6 3. d4 cxd4 4. Nxd4 Nf6 5. Nc3 e5 6. Ndb5 d6 7. Bg5
1. e4 c5 2. Nf3 Nc6 3. d4 cxd4 4. Nxd4 Nf6 5. Nc3 e5 6. Ndb5 d6 7. Nd5
1. e4 c5 2. Nf3 d6 3. Bb5+
1. e4 c5 2. Nf3 d6 3. d4 cxd4 4. Nxd4 Nf6 5. Nc3 Nc6
1. e4 c5 2. Nf3 d6 3. d4 cxd4 4. Nxd4 Nf6 5. Nc3 a6
1. e4 c5 2. Nf3 e6
1. e4 c5 2. Nf3 a6
1. e4 c5 2. Nf3 Nf6
1. e4 c5 2. Nf3 g6
//Открытые
1. e4 e5 2. Nc3 Nf6 3. g3
1. e4 e5 2. Nc3 Nf6 3. f4
1. e4 e5 2. Nc3 Nf6 3. Bc4
1. e4 e5 2. Nf3 Nf6 3. Nxe5
1. e4 e5 2. Nf3 Nf6 3. d4
1. e4 e5 2. Nf3 Nc6 3. d4 exd4 4. Nxd4
1. e4 e5 2. Nf3 Nc6 3. Bc4 Be7
1. e4 e5 2. Nf3 Nc6 3. Bc4 Bc5
1. e4 e5 2. Nf3 Nc6 3. Bb5 a6 4. Bxc6
1. e4 e5 2. Nf3 Nc6 3. Bb5 a6 4. Ba4 Nf6 5. O-O Nxe4
1. e4 e5 2. Nf3 Nc6 3. Bb5 a6 4. Ba4 Nf6 5. O-O Be7 6. Re1
1. e4 e5 2. Nf3 Nc6 3. Bb5 Nf6
//Филидор
1. e4 d6 2. d4 Nf6 3. Nc3 e5 4. dxe5 dxe5 5. Qxd8+ Kxd8
//Скандинавка
1. e4 d5
//Алехин
1. e4 Nf6
//Каро-Канн
1. e4 c6 2. Nc3 d5 3. Nf3 Bg4
1. e4 c6 2. Nc3 d5 3. Nf3 Nf6
1. e4 c6 2. Nc3 d5 3. Nf3 dxe4 4. Nxe4 Nf6
1. e4 c6 2. Nf3 d5 3. d3 dxe4 4. dxe4 Qxd1+ 5. Kxd1
1.e4 c6 2.d4 d5 3.f3
1.e4 c6 2.d4 d5 3.e5 c5
1.e4 c6 2.d4 d5 3.e5 Bf5
1.e4 c6 2.d4 d5 3.exd5 cxd5
1.e4 c6 2.d4 d5 3.Nc3 dxe4 4.Nxe4 Nf6
1.e4 c6 2.d4 d5 3.Nc3 dxe4 4.Nxe4 Bf5
1.e4 c6 2.d4 d5 3.Nc3 dxe4 4.Nxe4 Nd7
//Пирц
1. e4 d6 2. d4 Nf6 3. Bd3
1. e4 d6 2. d4 Nf6 3. Nc3 g6 4. Bg5
1. e4 d6 2. d4 Nf6 3. Nc3 g6 4. Be3 c6 5. Qd2
1. e4 d6 2. d4 Nf6 3. Nc3 g6 4. f4 Bg7 5. Nf3 c5 6. Bb5+
1. e4 d6 2. d4 Nf6 3. Nc3 g6 4. f4 Bg7 5. Nf3 c5 6. dxc5
1. e4 d6 2. d4 Nf6 3. Nc3 g6 4. f4 Bg7 5. Nf3 c5 6. 6. d5
1. e4 d6 2. d4 Nf6 3. Nc3 g6 4. f4 Bg7 5. Nf3 0-0
1. e4 d6 2. d4 Nf6 3. Nc3 g6 4. Nf3
//Нимцович
1.d4 Nf6 2.c4 e6 3.Nc3 Bb4
//Грюнфельд
1. d4 Nf6 2. c4 g6 3. Nc3 d5 4. cxd5 Nxd5 5. e4 Nxc3 6. bxc3 Bg7
//Эндшпиль 1. d4 d6
1. d4 d6 2. c4 e5 3. dxe5 dxe5 4. Qxd8+ Kxd8
//Славянская
1. d4 d5 2. c4 c6 3. Nf3 Nf6 4. Nc3 dxc4 5. a4 Bf5 6. Ne5
1. d4 d5 2. c4 c6 3. Nf3 Nf6 4. Nc3 dxc4 5. a4 Bf5 6. e3 e6 7. Bxc4 Bb4 8. O-O
1. d4 d5 2. c4 c6 3. Nf3 Nf6 4. Nc3 e6 5. e3 Nbd7 6. Qc2
1. d4 d5 2. c4 c6 3. Nf3 Nf6 4. Nc3 e6 5. e3 Nbd7 6. Bd3 dxc4 7. Bxc4 b5
1. d4 d5 2. c4 c6 3. Nf3 Nf6 4. Nc3 e6 5. e3 a6
1. d4 d5 2. c4 c6 3. Nc3 Nf6 4. Nf3 e6 5. Bg5 Nbd7 6. e3 Qa5
//Принятый ферзевый гамбит
1. d4 d5 2. c4 dxc4
//Лондон
1. d4 d5 2. Bf4
//Волжский гамбит
1. d4 Nf6 2. c4 c5 3. d5 b5 4. cxb5 a6 5. bxa6 g6 6. Nc3 Bg7 7. e4 O-O 8. a7 Rxa7
1. d4 Nf6 2. c4 c5 3. d5 b5 4. cxb5 a6 5. b6
1. d4 Nf6 2. c4 c5 3. d5 b5 4. cxb5 a6 5. e3
1. d4 Nf6 2. c4 c5 3. d5 b5 4. cxb5 a6 5. f3
1. d4 Nf6 2. c4 c5 3. d5 b5 4. cxb5 a6 5. Nc3
1. d4 Nf6 2. c4 c5 3. d5 b5 4. Nf3
//Будапешт
1. d4 Nf6 2. c4 e5 3. dxe5 Ng4
//Староиндийская
1. d4 Nf6 2. c4 g6 3. Nc3 Bg7 4. e4 d6 5. Nf3 O-O 6. Be2 e5
1. d4 Nf6 2. c4 g6 3. Nc3 Bg7 4. e4 d6 5. Nf3 O-O 6. Be2 Na6
//английское начало
1. c4 e5
`

func getOpenings() []string {
	var result []string
	var lines = strings.Split(openings, "\n")
	for _, line := range lines {
		if !(line == "" || strings.HasPrefix(line, "//")) {
			result = append(result, line)
		}
	}
	return result
}
