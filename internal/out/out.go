package out

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

type Printer struct {
	AsJSON bool
	Writer io.Writer
}

func New(asJSON bool) *Printer {
	return &Printer{
		AsJSON: asJSON,
		Writer: os.Stdout,
	}
}

func (p *Printer) Print(data interface{}, textHeaders []string, textRowFunc func(interface{}) [][]string) {
	if p.AsJSON {
		b, _ := json.MarshalIndent(data, "", "  ")
		fmt.Fprintln(p.Writer, string(b))
		return
	}

	w := tabwriter.NewWriter(p.Writer, 0, 0, 2, ' ', 0)
	if len(textHeaders) > 0 {
		for i, h := range textHeaders {
			fmt.Fprintf(w, "%s", h)
			if i < len(textHeaders)-1 {
				fmt.Fprint(w, "\t")
			}
		}
		fmt.Fprintln(w)
	}

	rows := textRowFunc(data)
	for _, row := range rows {
		for i, col := range row {
			fmt.Fprintf(w, "%s", col)
			if i < len(row)-1 {
				fmt.Fprint(w, "\t")
			}
		}
		fmt.Fprintln(w)
	}
	w.Flush()
}

func (p *Printer) Error(err error) {
	if p.AsJSON {
		b, _ := json.Marshal(map[string]string{"error": err.Error()})
		fmt.Fprintln(os.Stderr, string(b))
		return
	}
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
}
