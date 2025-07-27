package utils

import "github.com/zaf/g711"

func UlawToPcm(ulaw []byte) []int16 {
	pcm := make([]int16, len(ulaw))
	for index, data := range ulaw {
		pcm[index] = g711.DecodeUlawFrame(data)
	}

	return pcm
}
