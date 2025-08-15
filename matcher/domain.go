package matcher

import (
	"fmt"
	"strings"
)

type lableNode[V any] struct {
	children map[string]*lableNode[V]
	value    *V
}

type DomainMatcher[V any] struct {
	exactDomains map[string]*V
	root         *lableNode[V]
}

func NewDomainMatcher[V any]() *DomainMatcher[V] {
	return &DomainMatcher[V]{
		exactDomains: make(map[string]*V),
		root:         &lableNode[V]{children: make(map[string]*lableNode[V])},
	}
}

func SplitAndReverse(domain string) []string {
	parts := strings.Split(domain, ".")
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	return parts
}

func (m *DomainMatcher[V]) insertTrie(domain string, value V) {
	node := m.root
	lables := SplitAndReverse(domain)
	for _, lable := range lables {
		if node.children[lable] == nil {
			node.children[lable] = &lableNode[V]{children: make(map[string]*lableNode[V])}
		}
		node = node.children[lable]
	}
	node.value = &value
}

func (m *DomainMatcher[V]) Add(pattern string, value V) error {
	if !strings.Contains(pattern, ".") {
		return fmt.Errorf("invalid pattern: %s", pattern)
	} else if strings.HasPrefix(pattern, "*.") {
		m.insertTrie(pattern[2:], value)
	} else if strings.HasPrefix(pattern, "*") {
		domain := pattern[1:]
		m.exactDomains[domain] = &value
		m.insertTrie(domain, value)
	} else {
		m.exactDomains[pattern] = &value
	}
	return nil
}

func (m *DomainMatcher[V]) Find(domain string) *V {
	if value, ok := m.exactDomains[domain]; ok {
		return value
	}
	node := m.root
	lables := SplitAndReverse(domain)
	var value *V
	for _, lable := range lables {
		child, ok := node.children[lable]
		if !ok {
			break
		}
		node = child
		if node.value != nil {
			value = node.value
		}
	}
	return value
}