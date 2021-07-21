
# Ableton AlwaysHotKeys
Windows only.

_Made by Gianluca Elia for Daniel Sousa. Based on [schollz/osckeylogger](https://github.com/schollz/osckeylogger) (and that's why it's in Go)._

Sends alphanumeric keypresses (that is, a to z, 0 to 9 only) to Ableton even when Ableton is not in the foreground.



## install
Ok, first you need to install [git](https://git-scm.com/downloads) and [go](https://golang.org/doc/install).
Then:

```
git clone https://github.com/elgiano/abletonAlwaysHotKeys
cd abletonAlwaysHotKeys
go install -v
```

## usage

```
abletonAlwaysHotKeys
```

or, from the abletonAlwaysHotKeys folder

```
go run main.go
```
