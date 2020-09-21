// gqldoc: Generate documentation for GraphQL Schema.
//
// usage:
//     gqldoc -format output_format files ...
//
// output_format is one of:
//     graphql    GraphQL
//     gfm    GitHub Flavored Markdown
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/smasher164/gqldoc"
)

var format = flag.String("format", "", "output format of documentation")

func usage() {
	fmt.Fprint(os.Stderr, "gqldoc: Generate documentation for GraphQL Schema.\n")
	fmt.Fprint(os.Stderr, "\nusage:\n")
	fmt.Fprint(os.Stderr, "\tgqldoc -format output_format files ...\n")
	fmt.Fprint(os.Stderr, "\noutput_format is one of:\n")
	fmt.Fprint(os.Stderr, "\tgraphql\tGraphQL\n")
	fmt.Fprint(os.Stderr, "\tgfm\tGitHub Flavored Markdown\n")
	os.Exit(2)
}

func main() {
	log.SetPrefix("gqldoc: ")
	log.SetFlags(0)
	flag.Parse()
	if len(flag.Args()) == 0 {
		usage()
	}
	schema, err := gqldoc.ParseFiles(flag.Args())
	if err != nil {
		log.Fatal(err)
	}
	var render gqldoc.Formatter
	switch *format {
	case "graphql":
		render = gqldoc.FormatGraphQL
	case "gfm":
		render = gqldoc.FormatMarkdown
	default:
		usage()
	}
	if err := render(os.Stdout, schema); err != nil {
		log.Fatal(err)
	}
}
