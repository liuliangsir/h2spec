package h2spec

import (
	"fmt"
	"time"

	"github.com/summerwind/h2spec/config"
	"github.com/summerwind/h2spec/generic"
	"github.com/summerwind/h2spec/hpack"
	"github.com/summerwind/h2spec/http2"
	"github.com/summerwind/h2spec/log"
	"github.com/summerwind/h2spec/reporter"
	"github.com/summerwind/h2spec/spec"
)

func Run(c *config.Config) error {
	total := 0
	failed := false

	specs := []*spec.TestGroup{
		generic.Spec(),
		http2.Spec(),
		hpack.Spec(),
	}

	start := time.Now()
	for _, s := range specs {
		s.Test(c)

		if s.FailedCount > 0 {
			failed = true
		}

		total += s.FailedCount
		total += s.SkippedCount
		total += s.PassedCount
	}
	end := time.Now()
	d := end.Sub(start)

	if c.DryRun {
		return nil
	}

	if total == 0 {
		log.SetIndentLevel(0)
		log.Println("No matched tests found.")
		return nil
	}

	if failed {
		log.SetIndentLevel(0)
		reporter.FailedTests(specs)
	}

	log.SetIndentLevel(0)
	log.Println(fmt.Sprintf("Finished in %.4f seconds", d.Seconds()))
	reporter.Summary(specs)

	if c.JUnitReport != "" {
		err := reporter.JUnitReport(specs, c.JUnitReport)
		if err != nil {
			return err
		}
	}

	return nil
}
