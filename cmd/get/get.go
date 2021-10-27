// Package get implements the get command.
package get

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/alphasoc/flightsim/wisdom"
)

var Version string

var usage = `usage: flightsim get [flags] element:category

Available elements:

	%s

Available categories:

	%s

Available flags:
`

// printWelcome prints a basic welcome banner.
func printWelcome() {
	fmt.Printf(`
AlphaSOC Network Flight Simulator™ %s (https://github.com/alphasoc/flightsim)
The current time is %s
`, Version, time.Now().Format("02-Jan-06 15:04:05"))
}

// printGoodbye prints a parting message.
func printGoodbye() {
	fmt.Printf("\nAll done!\n")
}

// printMsg prints msg, decorated with an info string and the current date/time.
func printMsg(info string, msg string) {
	if msg == "" {
		return
	}
	fmt.Printf("%s [%s] %s\n", time.Now().Format("15:04:05"), info, msg)
}

// We only know how to get families for now.
var supportedElementsMap = map[string]bool{
	"families": true,
}

// supportedElements returns a slice of fetchable elements.
func supportedElements() []string {
	var elements []string
	for e := range supportedElementsMap {
		elements = append(elements, e)
	}
	return elements
}

// Taken from open-wisdom/server/entries/entries.go.
// TODO: would be good to fetch these also.
var supportedCategoriesMap = map[string]bool{
	"c2": true,
}

// supportedCategories returns a slice of fetchable categories.
func supportedCategories() []string {
	var categories []string
	for e := range supportedCategoriesMap {
		categories = append(categories, e)
	}
	return categories
}

// computeFormatStr returns a format string to be used with column printing.
func computeFormatStr(cols int) string {
	fmtStr := "\n"
	for i := 0; i < cols; i++ {
		fmtStr += "%v\t"
	}
	return fmtStr
}

// printCol prints elements in cols columns.
func printCol(elements []string, cols int) {
	w := new(tabwriter.Writer)
	// Min width, tab width, padding, pad char, flags.
	w.Init(os.Stdout, 8, 8, 1, '\t', 0)
	// Compute format string.
	fmtStr := computeFormatStr(cols)
	// Convert elements ([]string) to []interface{}.
	elementsToPrint := make([]interface{}, len(elements))
	for i, v := range elements {
		elementsToPrint[i] = v
	}
	// Print.
	i := 0
	lenElementsToPrint := len(elementsToPrint)
	for leftToPrint := lenElementsToPrint; leftToPrint > 0; {
		// We don't have enough elements left to print to satisfy the format string,
		// or the re-slice.  Thus, reset *cols and recompute the format string.
		if leftToPrint < cols {
			cols = leftToPrint
			fmtStr = computeFormatStr(cols)
		}
		fmt.Fprintf(w, fmtStr, elementsToPrint[i:i+cols]...)
		// Move by number of columns.
		i += cols
		leftToPrint = lenElementsToPrint - i
	}
	// Append a blank line.
	fmt.Fprintf(w, "%v", "\n\n")
	w.Flush()
}

// RunCmd runs the 'get' command and returns an error.
func RunCmd(args []string) error {
	printWelcome()
	// Mirrors look of run command.
	fmt.Println("")
	cmdLine := flag.NewFlagSet("get", flag.ExitOnError)
	// TODO: replace cols with -format (issue #45).
	// cols := cmdLine.Int("cols", 0, "print elements in number of columns")
	usageMsg := fmt.Sprintf(usage, strings.Join(supportedElements(), ", "), strings.Join(supportedCategories(), ", "))
	cmdLine.Usage = func() {
		fmt.Fprintf(cmdLine.Output(), usageMsg)
		cmdLine.PrintDefaults()
	}
	cmdLine.Parse(args)
	// Next arg should be element:category (ie. families:c2)
	toGet := cmdLine.Arg(0)
	if len(toGet) == 0 {
		return fmt.Errorf("nothing to get\n\n%v", usageMsg)
	}
	toGetArr := strings.Split(cmdLine.Arg(0), ":")
	if len(toGetArr) != 2 {
		return fmt.Errorf("unable to get '%v': invalid format", toGet)
	}
	elem := toGetArr[0]
	cat := toGetArr[1]
	// infoTag == "element:category" (ie. families:c2).  Mirrors the run command.
	infoTag := cmdLine.Arg(0)
	var elements []string
	var err error
	switch elem {
	case "families":
		printMsg(infoTag, fmt.Sprintf("Fetching %v %v", cat, elem))
		elements, err = wisdom.Families(cat)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unable to get '%v': unsupported element '%v'", toGet, elem)
	}
	// TODO: Leave quotes in for later cut/paste?
	for i := 0; i < len(elements); i++ {
		elements[i] = strings.Trim(elements[i], "\"")
	}
	// Default is to print in a single line.  Otherwise, column print.
	// if *cols <= 0 {
	// 	printMsg(infoTag, strings.Join(elements, ", "))
	// } else {
	// 	printCol(elements, *cols)
	// }
	printMsg(infoTag, strings.Join(elements, ", "))

	printMsg(infoTag, fmt.Sprintf("Fetched %v %v %v", len(elements), cat, elem))
	printGoodbye()
	return nil
}
