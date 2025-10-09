# fono

Terminal-based music player.

## Why?

Main reason behind the existence of this is to provide a fresh take on the 
original [fonograf](https://www.github.com/pes18fan/fonograf), to address
some of its shortcomings and add new features.

The most notable shortcoming of fonograf that this one fixes is the laggy,
flickering UI. This was caused because everything in fonograf was done on
one thread without any concurrency. However, this time, the audio unit and UI
unit operate separately, allowing for a much more responsive interface. This
is possible thanks to the ease of concurrency in Go, the language of choice
for this player.

I'll also be adding various new features to fono as I continue working on it.
While the [bubbletea](https://www.github.com/charmbracelet/bubbletea) library
allows for really nice interfaces in the terminal, I might consider making
the program a GUI in the future since a terminal can only do so much.
