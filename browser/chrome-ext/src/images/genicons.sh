#!/bin/sh

ico() {
  convert alkasir-icon-256x256.png "$@"
}

o=""
ico $o -resize 128x128! on.png
ico $o -resize 48x48! on48.png
ico $o -resize 16x16! on16.png

o="-fill grey -colorize 95%"
ico $o -resize 128x128! off.png
# ico $o -resize 48x48! off48.png

o="-fill gray  -colorize 100% -fill green -colorize 50%"
ico $o -resize 128x128! transported.png
# ico $o -resize 48x48! off48.png
# ico $o -resize 16x16! off16.png

o="-fill yellow  -colorize 30%"
ico $o -resize 128x128! warning.png
# ico $o -resize 48x48! off48.png
# ico $o -resize 16x16! off16.png

which optipng && optipng *.png
