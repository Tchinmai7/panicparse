package lib

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
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

func parseBucketHeader(bucket *stack.Bucket, multipleBuckets bool) string {
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

func stackLines(signature *stack.Signature, srcLen, pkgLen int) string {
	out := make([]string, len(signature.Stack.Calls))
	for i, line := range signature.Stack.Calls {
		out[i] = fmt.Sprintf("%-*s %-*s %s(%s)", pkgLen, line.Func.PkgName(), srcLen, formatCall(&line), line.Func.Name(), &line.Args)
	}
	if signature.Stack.Elided {
		out = append(out, "    (...)")
	}
	return strings.Join(out, "\n") + "\n"
}

func calcLengths(buckets []*stack.Bucket) (int, int) {
	srcLen := 0
	pkgLen := 0
	for _, bucket := range buckets {
		for _, line := range bucket.Signature.Stack.Calls {
			if l := len(formatCall(&line)); l > srcLen {
				srcLen = l
			}
			if l := len(line.Func.PkgName()); l > pkgLen {
				pkgLen = l
			}
		}
	}
	return srcLen, pkgLen
}

func ParsePanicString(stackTrace string) ([]string, error) {
	r := strings.NewReader(stackTrace)
	var junk bytes.Buffer
	writer := bufio.NewWriter(&junk)

	//writer would contain Junk after ParseDump
	ctx, err := stack.ParseDump(r, writer, true)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		return nil, errors.New("ctx is null")
	}
	stack.Augment(ctx.Goroutines)

	buckets := stack.Aggregate(ctx.Goroutines, stack.AnyPointer)
	multipleBuckets := len(buckets) > 1

	srcLen, pkgLen := calcLengths(buckets)
	out := make([]string, len(buckets))

	for i, bucket := range buckets {
		if bucket.First {
			header := parseBucketHeader(bucket, multipleBuckets)

			out[i] = fmt.Sprintf("%s%s", header, stackLines(&bucket.Signature, srcLen, pkgLen))
		}
	}

	return out, nil
}
