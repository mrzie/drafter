package utils

// type MaybeString struct {
// 	Value string
// }

// func (this MaybeString) MarshalJson() {

// }

// type MaybeInt struct {
// 	Int string
// }

func RunesIndexOf(s string, substring string) (index int) {
	lenSub := len(substring)
	lenSrc := len(s)
	if lenSub == 0 {
		return 0
	}
	if lenSrc == 0 {
		return -1
	}
	if lenSub == lenSrc {
		if s == substring {
			return 0
		}
		return -1
	}
	if lenSub > lenSrc {
		return -1
	}
	src := []rune(s)
	sub := []rune(substring)
	lengthSrc := len(src)
	lengthSub := len(sub)
	index = -1
	for i := 0; i < lengthSrc; i++ {
		for j := 0; j < lengthSub; j++ {
			if src[i+j] != sub[j] {
				break
			}
			if j+1 == lengthSub {
				index = i
				return
			}
		}
	}
	return
}
