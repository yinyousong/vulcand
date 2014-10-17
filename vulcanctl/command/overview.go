package command

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/mailgun/vulcand/Godeps/_workspace/src/github.com/buger/goterm"
	"github.com/mailgun/vulcand/backend"
)

func locationsOverview(locations []*backend.Location) string {
	t := goterm.NewTable(0, 10, 5, ' ', 0)
	fmt.Fprint(t, "Id\tHostname\tPath\tReqs/sec\t50ile [ms]\t99ile [ms]\tStatus codes %%%%\tNet. errors %%%%\n")

	if len(locations) == 0 {
		return t.String()
	}
	for _, l := range locations {
		locationOverview(t, l)
	}
	return t.String()
}

func endpointsOverview(endpoints []*backend.Endpoint) string {
	t := goterm.NewTable(0, 10, 5, ' ', 0)
	fmt.Fprint(t, "UpstreamId\tId\tUrl\tReqs/sec\t50ile [ms]\t99ile [ms]\tStatus codes %%%%\tNet. errors %%%%\tAnomalies\n")

	for _, e := range endpoints {
		endpointOverview(t, e)
	}
	return t.String()
}

func locationOverview(w io.Writer, l *backend.Location) {
	s := l.Stats

	fmt.Fprintf(w, "%s\t%s\t%s\t%0.1f\t%0.3f\t%0.3f\t%s\t%s\n",
		l.Id,
		l.Hostname,
		l.Path,
		s.RequestsPerSecond(),
		float64(s.LatencyBrackets[0].Value)/float64(time.Millisecond),
		float64(s.LatencyBrackets[len(s.LatencyBrackets)-1].Value)/float64(time.Millisecond),
		statusCodesToString(&s),
		errRateToString(s.NetErrorRate()),
	)
}

func endpointOverview(w io.Writer, e *backend.Endpoint) {
	s := e.Stats

	anomalies := ""
	if s.Verdict.IsBad {
		anomalies = fmt.Sprintf("%v", s.Verdict.Anomalies)
	}

	fmt.Fprintf(w, "%s\t%s\t%s\t%0.1f\t%0.3f\t%0.3f\t%s\t%s\t%s\n",
		e.UpstreamId,
		e.Id,
		e.Url,
		s.RequestsPerSecond(),
		float64(s.LatencyBrackets[0].Value)/float64(time.Millisecond),
		float64(s.LatencyBrackets[len(s.LatencyBrackets)-1].Value)/float64(time.Millisecond),
		statusCodesToString(&s),
		errRateToString(s.NetErrorRate()),
		anomalies)
}

func errRateToString(r float64) string {
	failRateS := fmt.Sprintf("%0.2f", r)
	if r != 0 {
		return goterm.Color(failRateS, goterm.RED)
	} else {
		return goterm.Color(failRateS, goterm.GREEN)
	}
}

func statusCodesToString(s *backend.RoundTripStats) string {
	if s.Counters.Total == 0 {
		return ""
	}

	sort.Sort(&codeSorter{codes: s.Counters.StatusCodes})

	codes := make([]string, 0, len(s.Counters.StatusCodes))
	for _, c := range s.Counters.StatusCodes {
		percent := 100 * (float64(c.Count) / float64(s.Counters.Total))
		out := fmt.Sprintf("%d: %0.2f", c.Code, percent)
		codes = append(codes, out)
	}

	return strings.Join(codes, ", ")
}

func getColor(code int) int {
	if code < 300 {
		return goterm.GREEN
	} else if code < 500 {
		return goterm.YELLOW
	}
	return goterm.RED
}

type codeSorter struct {
	codes []backend.StatusCode
}

func (c *codeSorter) Len() int {
	return len(c.codes)
}

func (c *codeSorter) Swap(i, j int) {
	c.codes[i], c.codes[j] = c.codes[j], c.codes[i]
}

func (c *codeSorter) Less(i, j int) bool {
	return c.codes[i].Code < c.codes[j].Code
}
