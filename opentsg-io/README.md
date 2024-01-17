# OpenTSG-io

## Description

opentsg-io contains the  encoding methods for custom image libraries, which are
not part of the golang/image library. This module is part of open tpg project,
but can be used independently for when you just need an image saved.

Thescurrent image files are:

- DPX files.
- Tiff files, which are part of the golang library, but this saves without any
  alpha channel.
- EXR files, there are alternatives avaiable.
- CSV representing the red blue green channels - MAY remove

At the moment no decoders are contained within in the repo, but this much change
with the needs of the project.

## Limtations

These functions are made to work with opentsg, so often use *image.NRGBA64
instead of the image.Image interface, because we do not handle the other images types. If you would
like to contribute to this to improve the options available to people who would
like to use this package please do.

**DPX:**

- 8 12 and 16 bits
- methods of encoding. little endian

**EXR:**

- only one method etc

## Visuals

Depending on what you are making, it can be a good idea to include screenshots
or even a video (you'll frequently see GIFs rather than actual videos). Tools
like ttygif can help, but check out Asciinema for a more sophisticated method.

## Installation

```sh
go get opentsg-io
```

## Usage

```go
import (
    image/image
    image/draw
    github.com/mmTristan/opentsg-modules/opentsg-io/dpx
)

func main() {

// make an image

// paint it green

//save it as a dpx


}
```

