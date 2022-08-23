package main

var openings = []string{
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
