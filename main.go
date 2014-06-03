// asciiart converts a picture into ascii art
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/nfnt/resize"
)

func errLog(s string, args ...interface{}) {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, s)
	} else {
		fmt.Fprintf(os.Stderr, s, args...)
	}
}

func main() {
	var (
		width                 = flag.Int("width", 80, "width of the output AA pic")
		alphabetFlag alphabet = alphabets["heuristic"]
	)
	flag.Var(&alphabetFlag, "alphabet", "which alphabet to use")

	flag.Usage = func() {
		errLog("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		alphas := []string{}
		for k, _ := range alphabets {
			alphas = append(alphas, k)
		}
		errLog("Recognized alphabets: %s\n", strings.Join(alphas, ", "))
	}

	flag.Parse()

	var (
		in  io.Reader
		out io.Writer
		err error
	)
	args := flag.Args()

	if len(args) > 0 {
		inf := args[0]

		switch inf[:4] {
		case "http":
			var resp *http.Response
			resp, err = http.Get(inf)
			if err != nil {
				errLog("Could not fetch HTTP resource %s : %s", inf, err)
				os.Exit(1)
			}
			if resp.StatusCode != 200 {
				errLog("No such location %s (%d)", inf, resp.StatusCode)
				os.Exit(1)
			}
			in = resp.Body
			defer func() {
				resp.Body.Close()
			}()

		default:
			in, err = os.Open(inf)
			if err != nil {
				errLog("Could not open input file %s: %s\n", inf, err)
				os.Exit(1)
			}

		}

	} else {
		in = os.Stdin
	}

	if len(args) > 1 {
		outf := args[1]
		out, err = os.Open(outf)
		if err != nil {
			errLog("Could not open output file %s: %s\n", outf, err)
			os.Exit(2)
		}
	} else {
		out = os.Stdout
	}

	img, _, err := image.Decode(in)
	if err != nil {
		errLog("Could not decode input image: %s\n", err)
		os.Exit(3)
	}

	thumb := resize.Resize(uint(*width), 0, img, resize.Bilinear)
	printTo(thumb, out, alphabetFlag)

}

type alphabet string

func (c *alphabet) String() string {
	for k, v := range alphabets {
		if *c == v {
			return k
		}
	}
	return "unknown"
}

func (a *alphabet) Set(value string) error {
	newA, ok := alphabets[value]
	if !ok {
		return fmt.Errorf("no such alphabet %s", value)
	}
	*a = newA
	return nil
}

// these map 0 to black and len(c) to white
var alphabets = map[string]alphabet{
	"heuristic": `#=$8Z7I\O?+:-,. `,
	"alternate": `@#8&o:*. `,
	"asciifi1":  `@GCLftli;:,. `,
	"asciifi2":  `#WMBRXVYIti+=;:,. `,
	"asciifi3":  `##XXxxx+++===---;;,,...  `,
}

func printTo(img image.Image, out io.Writer, mapping alphabet) {
	r := img.Bounds()
	gcm := color.GrayModel

	line := make([]byte, r.Dx())
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			g := gcm.Convert(img.At(x, y)).(color.Gray)
			c := mapping[int(float64(g.Y)/256.0*float64(len(mapping)))]
			line[x] = c
		}
		fmt.Fprintf(out, "%s\n", string(line))
	}
}
