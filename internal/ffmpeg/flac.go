package ffmpeg

import "audio_converter/internal/options"

// Converter options suitable for creating a FLAC audio file.
var FlacOptions = options.ConverterOptions{
	Codec:            "flac",
	InputExtensions:  InputExtensions,
	OutputExtensions: []string{".flac"},
}
