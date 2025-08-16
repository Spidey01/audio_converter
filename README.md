# Audio Converter

My tool bag for managing my audio collection.

## Dependencies

This requires a suitable version of ffmpeg to be installed and in path. ffmpeg
version 7.1.1 was current at time of writing. For compilation, the Go toolchain
is required; refer to go.mod for a compatible version.

## Build process

Just another go project. A Makefile is provided to easily run go install and go
test. By default, binaries are built in ./build and installed to
PREFIX=/usr/local, respecting DESTDIR.

```sh
    make
```

## Conversion Tools

There are several conversion tools.

| Tool     | Convert source to | Quality |
| -------- | - | - |
| to_aac   | M4A with AAC. | 256k VBR |
| to_flac  | Free Lossless Audio Codec. | Default |
| to_mp3   | MP3. | 320k VBR |

Each tool defaults to Stereo at 44.1 kHz sample rate and the above quality.
Flags can be used to override these if desired. Cover art and metadata will
typically be converted but milage may vary.

The default settings generally favor high quality then compatibility. For
example, just about everything can handle 44.1 kHz (CD quality) and 256k is more
than enough bits for AAC-LC handle stereo audio with pretty high compatibility.

Additionally, there are tools for more specific purposes.

| Script  | Comment |
| ------- | ------- |
| export_audio_tree | Convert a directory tree. Useful for exporting libraries and albums. |
| extract_coverart  | Extracts the cover art with optional scaling and format conversion. |


### Example of Converting Single Files

```sh
to_aac  input.flac output.m4a
to_flac input.wav output.flac
to_mp3  input.flac output.mp3
```

Use `-h` to get more detailed usage. Options are mainly to change the quality
settings like bitrate and sample rate. Each of the tools is very similar. E.g.,
to_flac doesn't take a bitrate flag, but to_aac does.

### Example of Converting a Tree

```sh
export_audio_tree -f flac ./in ./out
```

This would replicate the library of files in in to out using the associated
to_aac script. E.g., "in/album/song.flac" would become "out/album/song.m4a." By
default, unknown files are copied, so that ancillery files will be exported.

Use `-h` option for more details. Options cover most things.

### Example of Extracting Cover Art

```sh
extract_coverart -s 500x500 input.m4a cover.jpg
```

Would extract the cover art from the m4a file, scale it to 500 by 500 pixels, and store it in cover.jpg.

## Suggested Third Party Programs

Tools that I've found very helpful:

- [Mp3tag](https://mp3tag.app) for editing tags, embedding cover art, and renaming files.
- [Audacity](https://www.audacityteam.org) for studying audio.
