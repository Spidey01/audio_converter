// SPDX-License-Identifier: Zlib
// Copyright 2025, Terry M. Poulin.
package ffmpeg

import "audio_converter/internal/options"

// Converter options suitable for creating an MP4 audio file.
var AacOptions = options.ConverterOptions{
	BitRate:          "256k",
	Codec:            "aac",
	InputExtensions:  InputExtensions,
	OutputExtensions: []string{".m4a", ".m4r"},
}
