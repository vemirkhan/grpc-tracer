// Package tracetagger enriches spans with additional tags based on configurable
// matching rules.
//
// Rules can match on service name prefix, method substring, or a combination of
// both. When a span satisfies a rule's conditions all tags defined in that rule
// are written onto the span and the updated span is persisted back to the
// underlying TraceStore.
//
// Example usage:
//
//	store := storage.NewTraceStore()
//	tgr, err := tracetagger.New(store)
//	if err != nil {
//		log.Fatal(err)
//	}
//	tgr.AddRule(tracetagger.Rule{
//		ServicePrefix: "payment",
//		Tags: map[string]string{"team": "finance", "critical": "true"},
//	})
//	n, err := tgr.TagSpan(span)
package tracetagger
