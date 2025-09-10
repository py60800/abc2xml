// test
package main

import (
	"os"
	"strings"

	"path"

	"github.com/py60800/abc2xml"
)

func main() {
	src := "test.abc"
	if len(os.Args) > 1 {
		src = os.Args[1]
	}
	data, _ := os.ReadFile(src)

	parser := abc2xml.Abc2xmlNew()
	parser.SetDivisions(120)
	xml, _ := parser.Run(string(data))

	ext := path.Ext(path.Base(src))
	dest := strings.TrimSuffix(src, ext) + ".xml"
	res, _ := os.Create(dest)
	res.WriteString(xml)
	res.Close()

}
