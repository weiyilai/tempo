package traceql

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSpansetClone(t *testing.T) {
	ss := []*Spanset{
		{
			Spans: []Span{
				&mockSpan{
					id:                 []byte{0x01},
					startTimeUnixNanos: 3,
					durationNanos:      2,
				},
			},
			Scalar:             NewStaticFloat(3.2),
			TraceID:            []byte{0x02},
			RootSpanName:       "a",
			RootServiceName:    "b",
			StartTimeUnixNanos: 1,
			DurationNanos:      5,
			Attributes:         []*SpansetAttribute{{Name: "foo", Val: NewStaticString("bar")}},
		},
		{
			Spans: []Span{
				&mockSpan{
					id:                 []byte{0x01},
					startTimeUnixNanos: 3,
					durationNanos:      2,
				},
			},
			Scalar:             NewStaticFloat(3.2),
			TraceID:            []byte{0x02},
			RootSpanName:       "a",
			RootServiceName:    "b",
			StartTimeUnixNanos: 1,
			DurationNanos:      5,
		},
	}

	for _, s := range ss {
		require.True(t, reflect.DeepEqual(s, s.clone()))
	}
}

func TestMetaConditionsWithout(t *testing.T) {
	tcs := []struct {
		query  string
		expect []Condition
	}{
		{
			// No meta fields present in query, all are selected.
			query:  "{ status=error}",
			expect: SearchMetaConditions(),
		},
		{
			// Service name, span name are able to be reused
			query: "{ rootServiceName = `foo` && rootName = `bar`}",
			expect: []Condition{
				{Attribute: NewIntrinsic(IntrinsicTraceDuration), Op: OpNone},
				{Attribute: NewIntrinsic(IntrinsicTraceID), Op: OpNone},
				{Attribute: NewIntrinsic(IntrinsicTraceStartTime), Op: OpNone},
				{Attribute: NewIntrinsic(IntrinsicSpanID), Op: OpNone},
				{Attribute: NewIntrinsic(IntrinsicSpanStartTime), Op: OpNone},
				{Attribute: NewIntrinsic(IntrinsicDuration), Op: OpNone},
				{Attribute: NewIntrinsic(IntrinsicServiceStats), Op: OpNone},
			},
		},
		{
			// Duration is the only one able to be reused because it has no filtering
			query: "{ rootServiceName = `foo` && rootName = `bar`} | avg(duration) > 1s",
			expect: []Condition{
				{Attribute: NewIntrinsic(IntrinsicTraceRootService), Op: OpNone},
				{Attribute: NewIntrinsic(IntrinsicTraceRootSpan), Op: OpNone},
				{Attribute: NewIntrinsic(IntrinsicTraceDuration), Op: OpNone},
				{Attribute: NewIntrinsic(IntrinsicTraceID), Op: OpNone},
				{Attribute: NewIntrinsic(IntrinsicTraceStartTime), Op: OpNone},
				{Attribute: NewIntrinsic(IntrinsicSpanID), Op: OpNone},
				{Attribute: NewIntrinsic(IntrinsicSpanStartTime), Op: OpNone},
				{Attribute: NewIntrinsic(IntrinsicServiceStats), Op: OpNone},
			},
		},
		{
			// None are reused because the values are filtered and allConditions=false
			query:  "{ rootServiceName = `foo` || rootName = `bar`}",
			expect: SearchMetaConditions(),
		},
	}

	for _, tc := range tcs {
		req, _ := ExtractFetchSpansRequest(tc.query)
		require.Equal(t, tc.expect, SearchMetaConditionsWithout(req.Conditions, req.AllConditions))
	}
}
