// SPDX-License-Identifier: Zlib
// Copyright 2025, Terry M. Poulin.
package ffmpeg

import "audio_converter/internal/options"

// Converter options suitable for creating an MP3 audio file.
var Mp3Options = &options.ConverterOptions{
	BitRate:          "320k",
	Codec:            "libmp3lame",
	InputExtensions:  InputExtensions,
	OutputExtensions: []string{".mp3"},
}
