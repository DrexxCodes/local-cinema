package scanner

import "sync"

const maxIssues = 20

// issueList collects the problems a scan ran into, so the Library tab can show them.
type issueList struct {
	mu   sync.Mutex
	list []string
}

func (l *issueList) reset() {
	l.mu.Lock()
	l.list = nil
	l.mu.Unlock()
}

// add keeps the first maxIssues messages and drops the rest.
func (l *issueList) add(msg string) {
	l.mu.Lock()
	if len(l.list) < maxIssues {
		l.list = append(l.list, msg)
	}
	l.mu.Unlock()
}

// all returns a copy, never nil, so it encodes as [] in JSON.
func (l *issueList) all() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]string{}, l.list...)
}
