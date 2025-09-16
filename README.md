# go-abc2xml

go-abc2xml is intented to convert music written using ABC music notation to musicxml format.

Go-abc2xml is pure go partial replacement for python based [abc2xml](https://wim.vree.org/svgParse/abc2xml.html).

The original abc2xml does a very good job and can be invoked pretty easily from a go program but the end users of another program I wrote, found difficult to install python3 and libraries required by abc2xml in their environment. This is why I decided to write an alternative to the original abc2xml.

## limitations

go-abc2xml should be able to convert most of the traditionnal tunes available on the Internet. Particularly those available in [thesession](thesession.org).

Main limitations are:
- no support for fancy ABC features
- no support for multiple voices
- only basic n-tuplets are supported
- limited support for decorations

## testing

As of now, testing has been done by comparing the output of the original abc2xml to the output of go-abc2xml and by judging the result of the conversion using MuseScore4.

## future

Beside bug fixes, I have no intention to provide a full support to ABC notation however I may improve the the support for traditionnal tunes if I need it or if there is a demand for it.
IMHO musicxml is a far more mature music notation format and should be prefered whenever possible for complex tunes. ABC music notation is great for simple traditionnal tunes and for the huge amount of tunes available on the Internet.
