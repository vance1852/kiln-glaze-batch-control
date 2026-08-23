package audit

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

var errFailed = errors.New("persist failed")

func validEvent(action, outcome string) Event {
	return Event{RequestID: "req-1", ObjectType: "task", ObjectID: "s-1", Action: action, Outcome: outcome, CreatedAt: time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)}
}

func TestChainChangesHeadAndPreservesSequence(t *testing.T) {
	var chain Chain
	first, err := chain.Append(validEvent("create", "success"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := chain.Append(validEvent("complete", "success"))
	if err != nil {
		t.Fatal(err)
	}
	if first == second || chain.Head() != second {
		t.Fatalf("head=%s first=%s second=%s", chain.Head(), first, second)
	}
}

func TestFilterEventsMatchesAllProvidedFields(t *testing.T) {
	events := []Event{validEvent("create", "success"), validEvent("review", "success"), validEvent("review", "failure")}
	filtered := FilterEvents(events, Filter{Action: "review", Outcome: "success"})
	if len(filtered) != 1 || filtered[0].Action != "review" {
		t.Fatalf("filtered=%+v", filtered)
	}
	if len(FilterEvents(events, Filter{ObjectID: "missing"})) != 0 {
		t.Fatal("unmatched object returned")
	}
}

func TestChainAdvanceCommitsOnlyOnSuccess(t *testing.T) {
	var chain Chain
	failing := validEvent("create", "success")
	head, err := chain.Advance(failing, func(string) error { return errFailed })
	if !errors.Is(err, errFailed) || head != "" {
		t.Fatalf("head=%q err=%v", head, err)
	}
	if chain.Head() != "" {
		t.Fatalf("head advanced after failed persist: %s", chain.Head())
	}
}

func TestChainConcurrentAdvanceProducesUniqueHeads(t *testing.T) {
	const goroutines = 64
	var chain Chain
	heads := make([]string, goroutines)
	start := make(chan struct{})
	var done sync.WaitGroup
	done.Add(goroutines)
	for i := range goroutines {
		go func(i int) {
			defer done.Done()
			event := validEvent("create", "success")
			event.RequestID = fmt.Sprintf("req-%d", i)
			<-start
			head, err := chain.Advance(event, func(h string) error { heads[i] = h; return nil })
			if err != nil || head != heads[i] || head == "" {
				t.Errorf("goroutine %d: head=%q err=%v", i, head, err)
			}
		}(i)
	}
	close(start)
	done.Wait()
	seen := make(map[string]struct{}, goroutines)
	for _, head := range heads {
		if _, dup := seen[head]; dup {
			t.Fatalf("duplicate chain head under concurrency: %s", head)
		}
		seen[head] = struct{}{}
	}
}

func TestExportCSVHasHeaderAndRows(t *testing.T) {
	var output bytes.Buffer
	if err := ExportCSV(&output, []Event{validEvent("create", "success")}); err != nil {
		t.Fatal(err)
	}
	text := output.String()
	if !strings.Contains(text, "request_id,object_type") || strings.Count(text, "req-1") != 1 {
		t.Fatalf("csv=%s", text)
	}
}
