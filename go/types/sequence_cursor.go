// Copyright 2016 Attic Labs, Inc. All rights reserved.
// Licensed under the Apache License, version 2.0:
// http://www.apache.org/licenses/LICENSE-2.0

package types

import "github.com/attic-labs/noms/go/d"

// sequenceCursor explores a tree of sequence items.
type sequenceCursor struct {
	parent *sequenceCursor
	seq    sequence
	idx    int
}

func newSequenceCursor(parent *sequenceCursor, seq sequence, idx int) *sequenceCursor {
	if idx < 0 {
		idx += seq.seqLen()
		d.Chk.True(idx >= 0)
	}

	return &sequenceCursor{parent, seq, idx}
}

func (cur *sequenceCursor) length() int {
	return cur.seq.seqLen()
}

func (cur *sequenceCursor) getItem(idx int) sequenceItem {
	return cur.seq.getItem(idx)
}

func (cur *sequenceCursor) sync() {
	d.Chk.True(cur.parent != nil)
	cur.seq = cur.parent.getChildSequence()
}

func (cur *sequenceCursor) getChildSequence() sequence {
	if ms, ok := cur.seq.(metaSequence); ok {
		return ms.getChildSequence(cur.idx)
	}
	return nil
}

// Returns the value the cursor refers to. Fails an assertion if the cursor doesn't point to a value.
func (cur *sequenceCursor) current() sequenceItem {
	d.Chk.True(cur.valid())
	return cur.getItem(cur.idx)
}

func (cur *sequenceCursor) valid() bool {
	return cur.idx >= 0 && cur.idx < cur.length()
}

func (cur *sequenceCursor) depth() int {
	if nil != cur.parent {
		return 1 + cur.parent.depth()
	}
	return 1
}

func (cur *sequenceCursor) indexInChunk() int {
	return cur.idx
}

func (cur *sequenceCursor) advance() bool {
	return cur.advanceMaybeAllowPastEnd(true)
}

func (cur *sequenceCursor) advanceMaybeAllowPastEnd(allowPastEnd bool) bool {
	if cur.idx < cur.length()-1 {
		cur.idx++
		return true
	}
	if cur.idx == cur.length() {
		return false
	}
	if cur.parent != nil && cur.parent.advanceMaybeAllowPastEnd(false) {
		cur.sync()
		cur.idx = 0
		return true
	}
	if allowPastEnd {
		cur.idx++
	}
	return false
}

func (cur *sequenceCursor) retreat() bool {
	return cur.retreatMaybeAllowBeforeStart(true)
}

func (cur *sequenceCursor) retreatMaybeAllowBeforeStart(allowBeforeStart bool) bool {
	if cur.idx > 0 {
		cur.idx--
		return true
	}
	if cur.idx == -1 {
		return false
	}
	d.Chk.True(0 == cur.idx)
	if cur.parent != nil && cur.parent.retreatMaybeAllowBeforeStart(false) {
		cur.sync()
		cur.idx = cur.length() - 1
		return true
	}
	if allowBeforeStart {
		cur.idx--
	}
	return false
}

func (cur *sequenceCursor) clone() *sequenceCursor {
	var parent *sequenceCursor
	if cur.parent != nil {
		parent = cur.parent.clone()
	}
	return &sequenceCursor{parent, cur.seq, cur.idx}
}

type cursorIterCallback func(item interface{}) bool

func (cur *sequenceCursor) iter(cb cursorIterCallback) {
	for cur.valid() && !cb(cur.getItem(cur.idx)) {
		cur.advance()
	}
}
