package lib

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/Tchinmai7/panicparse/stack"
)

func formatCall(c *stack.Call) string {
	return fmt.Sprintf("%s:%d", c.SrcName(), c.Line)
}

func createdByString(s *stack.Signature) string {
	created := s.CreatedBy.Func.PkgDotName()

	if created == "" {
		return ""
	}
	return created + " @ " + formatCall(&s.CreatedBy)
}

func parseBucketHeader(bucket stack.Bucket, multipleBuckets bool) string {
	pf := basePath
	extra := ""
	if s := bucket.SleepString(); s != "" {
		extra += " [" + s + "]"
	}
	if bucket.Locked {
		extra += " [locked]"
	}
	if c := createdByString(&bucket.Signature); c != "" {
		extra += " [Created by " + c + "]"
	}
	return fmt.Sprintf("%d: %s%s\n", len(bucket.IDs), bucket.State, extra)
}

func stackLines(signature *stack.Signature) string {
	out := make([]string, len(signature.Stack.Calls))
	for i, line := range signature.Stack.Calls {
		out[i] = fmt.Sprintf("%s %s %s(%s)", line.Func.PkgName(), formatCall(line), line.Func.Name(), &line.Args)
	}
	if signature.Stack.Elided {
		out = append(out, "    (...)")
	}
	return strings.Join(out, "\n") + "\n"
}

func ParsePanicString(stackTrace string, guessPaths bool) string {
	r := strings.NewReader(stackTrace)
	var writer bytes.Buffer
	//writer would contain Junk after ParseDump
	ctx, err := stack.ParseDump(r, writer, guessPaths)
	writer.Reset()
	buckets := stack.Aggregate(ctx.Goroutines)
	multipleBuckets := len(buckets) > 1
	for _, bucket := range buckets {
		header := parseBucketHeader(bucket, multipleBuckets)
		_, _ = io.WriteString(writer, header)
		_, _ = io.WriteString(writer, stackLines(&bucket.Signature))
	}
	return writer.String()
}
