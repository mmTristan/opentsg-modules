# ACES

ACES is custom image library for [aces images](https://www.oscars.org/science-technology/sci-tech-projects/aces),
it is currently under development.
It use native go float 32
for pixel values instead of float16 as go does not currently support it.
These are converted to float16 when saved in the EXR files.

ACES image are compatible with the go image library.
