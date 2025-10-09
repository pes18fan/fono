# fono

Terminal-based music player.

## Usage

Just invoke it on the terminal:

```
fono
```

It'll show you a filepicker where you can choose an audio file from the current
directory. Currently, `mp3`, `wav`, `flac` and `ogg` files are supported. On
choosing a file it will take you to the player screen where you can pause/play
the file, change to another one or quit.

## Building

You'll need the Go compiler.

Clone the repo:

```
git clone https://github.com/pes18fan/fono.git
```

Go to the `fono` directory, then get the required libraries and build it to get
the `fono` binary by running

```
go mod tidy
go build ./cmd
```

Have fun!

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
