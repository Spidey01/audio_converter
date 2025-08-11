package main

import (
	"audio_converter/internal/ffmpeg"
)

func main() {
	ffmpeg.ConvertMain(ffmpeg.Mp3Options)
}
