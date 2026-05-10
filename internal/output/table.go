package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

func PrintTable(data [][]string, header []string) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if len(header) > 0 {
		fmt.Fprint(w, strings.Join(header, "\t"))
		fmt.Fprintln(w)
		for i := range header {
			fmt.Fprint(w, strings.Repeat("-", len(header[i]))+"\t")
		}
		fmt.Fprintln(w)
	}
	for _, row := range data {
		fmt.Fprint(w, strings.Join(row, "\t"))
		fmt.Fprintln(w)
	}
	return w.Flush()
}

func PrintJSON(data interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}