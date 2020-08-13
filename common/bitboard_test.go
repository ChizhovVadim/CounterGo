package common

import (
	"fmt"
	"math/bits"
	"testing"
)

func TestBitboard(t *testing.T) {
	type args struct {
		value uint64
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// test cases.
		{"A", args{FileAMask}, true},
		{"B", args{FileBMask}, true},
		{"C", args{FileCMask}, true},
		{"D", args{FileDMask}, true},
		{"E", args{FileEMask}, true},
		{"F", args{FileFMask}, true},
		{"G", args{FileGMask}, true},
		{"H", args{FileHMask}, true},
		{"1", args{Rank1Mask}, true},
		{"2", args{Rank2Mask}, true},
		{"3", args{Rank3Mask}, true},
		{"4", args{Rank4Mask}, true},
		{"5", args{Rank5Mask}, true},
		{"6", args{Rank6Mask}, true},
		{"7", args{Rank7Mask}, true},
		{"8", args{Rank8Mask}, true},
		{"bishop", args{0x0004085000500800}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Printf("OnesCount64(%064b) = %d\n", tt.args.value, bits.OnesCount64(tt.args.value))
			if !tt.want {
				t.Errorf("Bitbord want %v", tt.want)
			}
		})
	}
}

func TestMoreThanOne(t *testing.T) {
	type args struct {
		value uint64
	}
	tests := []struct {
		name    string
		args    args
		version int
		want    bool
	}{
		// test cases.
		{"zero", args{0}, 1, false},
		{"one", args{1}, 1, false},
		{"far one", args{1 << 5}, 1, false},
		{"farer one", args{1 << 60}, 1, false},
		{"two ones", args{3}, 1, true},
		{"two ones apart", args{1<<6 | 1<<25}, 1, true},
		{"three ones apart", args{1<<6 | 1<<25 | 1<<36}, 1, true},
		{"zero", args{0}, 2, false},
		{"one", args{1}, 2, false},
		{"far one", args{1 << 5}, 2, false},
		{"farther one", args{1 << 60}, 2, false},
		{"two ones", args{3}, 2, true},
		{"two ones apart", args{1<<6 | 1<<25}, 2, true},
		{"three ones apart", args{1<<6 | 1<<25 | 1<<36}, 2, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			switch tt.version {
			case 1:
				got = MoreThanOne(tt.args.value)
			case 2:
				got = OldMoreThanOne(tt.args.value)
			}
			fmt.Printf("%v OnesCount64(%064b) = %d\n %064b\n", got, tt.args.value, bits.OnesCount64(tt.args.value), (tt.args.value - 1))
			if got != tt.want {
				switch tt.version {
				case 1:
					t.Errorf("MoreThanOne() = %v, want %v", got, tt.want)
				case 2:
					t.Errorf("OldMoreThanOne() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// benchmark
func BenchmarkMoreThanOne(b *testing.B) {
	type args struct {
		value uint64
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// test cases.
		{"zero", args{0}, false},
		{"one", args{1}, false},
		{"far one", args{1 << 5}, false},
		{"farer one", args{1 << 60}, false},
		{"two ones", args{3}, true},
		{"two ones apart", args{1<<6 | 1<<25}, true},
		{"three ones apart", args{1<<6 | 1<<25 | 1<<36}, true},
		{"four ones apart", args{1<<6 | 1<<25 | 1<<36 | 1<<42}, true},
	}

	b.Run("MoreThanOne", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			for _, tt := range tests {
				if got := MoreThanOne(tt.args.value); got != tt.want {
					b.Errorf("MoreThanOne() = %v, want %v", got, tt.want)
				}
			}
		}
	})

	b.Run("NewMoreThanOne", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			for _, tt := range tests {
				if got := OldMoreThanOne(tt.args.value); got != tt.want {
					b.Errorf("NewMoreThanOne() = %v, want %v", got, tt.want)
				}
			}
		}
	})

}

func TestFirstOne(t *testing.T) {
	type args struct {
		value uint64
	}
	tests := []struct {
		name string
		args args
	}{
		// test cases.
		{"A", args{FileAMask}},
		{"B", args{FileBMask}},
		{"C", args{FileCMask}},
		{"D", args{FileDMask}},
		{"E", args{FileEMask}},
		{"F", args{FileFMask}},
		{"G", args{FileGMask}},
		{"H", args{FileHMask}},
		{"1", args{Rank1Mask}},
		{"2", args{Rank2Mask}},
		{"3", args{Rank3Mask}},
		{"4", args{Rank4Mask}},
		{"5", args{Rank5Mask}},
		{"6", args{Rank6Mask}},
		{"7", args{Rank7Mask}},
		{"8", args{Rank8Mask}},
		{"bishop", args{0x0004085000500800}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Printf("OnesCount64(%064b) = %d\nFirstOne = %d trailing0 %d\n", tt.args.value, bits.OnesCount64(tt.args.value), FirstOne(tt.args.value), OldFirstOne(tt.args.value))
			if FirstOne(tt.args.value) != OldFirstOne(tt.args.value) {
				t.Errorf("FirstOne want %d trailinngZero sais %d", FirstOne(tt.args.value), OldFirstOne(tt.args.value))
			}
		})
	}
}

/*
BenchmarkFirstOne
BenchmarkFirstOne/FirstOne
BenchmarkFirstOne/FirstOne-8         	78964704	        13.9 ns/op
BenchmarkFirstOne/OldFirstOne
BenchmarkFirstOne/OldFirstOne-8      	52768682	        21.1 ns/op
*/
func BenchmarkFirstOne(b *testing.B) {
	type args struct {
		value uint64
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// test cases.
		{"A", args{FileAMask}, true},
		{"B", args{FileBMask}, true},
		{"C", args{FileCMask}, true},
		{"D", args{FileDMask}, true},
		{"E", args{FileEMask}, true},
		{"F", args{FileFMask}, true},
		{"G", args{FileGMask}, true},
		{"H", args{FileHMask}, true},
		{"1", args{Rank1Mask}, true},
		{"2", args{Rank2Mask}, true},
		{"3", args{Rank3Mask}, true},
		{"4", args{Rank4Mask}, true},
		{"5", args{Rank5Mask}, true},
		{"6", args{Rank6Mask}, true},
		{"7", args{Rank7Mask}, true},
		{"8", args{Rank8Mask}, true},
		{"bishop", args{0x0004085000500800}, true},
	}

	b.Run("FirstOne", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			for _, tt := range tests {
				got := FirstOne(tt.args.value)
				if !tt.want {
					b.Errorf("FirstOne() = %d, want %v", got, tt.want)
				}
			}
		}
	})

	b.Run("OldFirstOne", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			for _, tt := range tests {
				got := OldFirstOne(tt.args.value)
				if !tt.want {
					b.Errorf("OldFirstOne() = %d, want %v", got, tt.want)
				}
			}
		}
	})

}
