// Copied from https://github.com/smacker/go-tree-sitter/blob/7d35f700adf0df32d3e7c80c56f7180376c1c11d/bindings.go

package sitter

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"unsafe"

	"github.com/yourbase/treesitter/internal/lang"
	C "github.com/yourbase/treesitter/internal/lib"
	"modernc.org/libc"
	"modernc.org/libc/sys/types"
)

// maintain a map of read functions that can be called from C
var readFuncs = &readFuncsMap{funcs: make(map[int]ReadFunc)}

// Parse is a shortcut for parsing bytes of source code,
// returns root node
func Parse(content []byte, lang *Language) *Node {
	p := NewParser()
	p.SetLanguage(lang)
	return p.Parse(nil, content).RootNode()
}

// Parser produces concrete syntax tree based on source code using Language
type Parser struct {
	tls      *libc.TLS
	isClosed bool
	c        uintptr
}

// NewParser creates new Parser
func NewParser() *Parser {
	tls := libc.NewTLS()
	p := &Parser{tls: tls, c: C.Xts_parser_new(tls)}
	runtime.SetFinalizer(p, (*Parser).Close)
	return p
}

// SetLanguage assignes Language to a parser
func (p *Parser) SetLanguage(l *Language) {
	C.Xts_parser_set_language(p.tls, p.c, lang.LanguagePtr(l))
}

// ReadFunc is a function to retrieve a chunk of text at a given byte offset and (row, column) position
// it should return nil to indicate the end of the document
type ReadFunc func(offset uint32, position Point) []byte

// InputEncoding is a encoding of the text to parse
type InputEncoding int

const (
	InputEncodingUTF8 InputEncoding = iota
	InputEncodingUTF16
)

// Input defines parameters for parse method
type Input struct {
	Read     ReadFunc
	Encoding InputEncoding
}

// Parse produces new Tree from content using old tree
func (p *Parser) Parse(oldTree *Tree, content []byte) *Tree {
	var BaseTree uintptr
	if oldTree != nil {
		BaseTree = oldTree.c
	}

	//input := libc.CBytes(content)
	input := uintptr(0)
	BaseTree = C.Xts_parser_parse_string(p.tls, p.c, BaseTree, input, uint32(len(content)))
	libc.Xfree(p.tls, input)

	return p.newTree(BaseTree)
}

// TODO(light)
// ParseInput produces new Tree by reading from a callback defined in input
// it is useful if your data is stored in specialized data structure
// as it will avoid copying the data into []bytes
// and faster access to edited part of the data
//
// func (p *Parser) ParseInput(oldTree *Tree, input Input) *Tree {
// 	var BaseTree uintptr
// 	if oldTree != nil {
// 		BaseTree = oldTree.c
// 	}

// 	funcID := readFuncs.register(input.Read)
// 	BaseTree = C.call_ts_parser_parse(p.c, BaseTree, C.int(funcID), C.TSInputEncoding(input.Encoding))
// 	readFuncs.unregister(funcID)

// 	return p.newTree(BaseTree)
// }

// OperationLimit returns the duration in microseconds that parsing is allowed to take
func (p *Parser) OperationLimit() int {
	return int(C.Xts_parser_timeout_micros(p.tls, p.c))
}

// SetOperationLimit limits the maximum duration in microseconds that parsing should be allowed to take before halting
func (p *Parser) SetOperationLimit(limit int) {
	C.Xts_parser_set_timeout_micros(p.tls, p.c, uint64(limit))
}

// Reset causes the parser to parse from scratch on the next call to parse, instead of resuming
// so that it sees the changes to the beginning of the source code.
func (p *Parser) Reset() {
	C.Xts_parser_reset(p.tls, p.c)
}

// SetIncludedRanges sets text ranges of a file
func (p *Parser) SetIncludedRanges(ranges []Range) {
	const rangeSize = types.Size_t(unsafe.Sizeof(C.TSRange{}))
	cranges := libc.Xmalloc(p.tls, types.Size_t(len(ranges))*rangeSize)
	defer libc.Xfree(p.tls, cranges)
	for i, r := range ranges {
		ptr := (*C.TSRange)(unsafe.Pointer(cranges + uintptr(i)*uintptr(rangeSize)))
		*ptr = C.TSRange{
			Start_point: C.TSPoint{
				Row:    uint32(r.StartPoint.Row),
				Column: uint32(r.StartPoint.Column),
			},
			End_point: C.TSPoint{
				Row:    uint32(r.EndPoint.Row),
				Column: uint32(r.EndPoint.Column),
			},
			Start_byte: uint32(r.StartByte),
			End_byte:   uint32(r.EndByte),
		}
	}
	C.Xts_parser_set_included_ranges(p.tls, p.c, cranges, uint32(len(ranges)))
}

// Debug enables debug output to stderr
func (p *Parser) Debug() {
	panic("TODO(light)")
	// logger := C.stderr_logger_new(true)
	// C.Xts_parser_set_logger(p.c, logger)
}

// Close should be called to ensure that all the memory used by the parse is freed.
func (p *Parser) Close() {
	if !p.isClosed {
		C.Xts_parser_delete(p.tls, p.c)
		p.tls.Close()
	}

	p.isClosed = true
}

type Point struct {
	Row    uint32
	Column uint32
}

type Range struct {
	StartPoint Point
	EndPoint   Point
	StartByte  uint32
	EndByte    uint32
}

// we use cache for nodes on normal tree object
// it prevent run of SetFinalizer as it introduces cycle
// we can workaround it using separate object
// for details see: https://github.com/golang/go/issues/7358#issuecomment-66091558
type BaseTree struct {
	tls      *libc.TLS
	c        uintptr
	isClosed bool
}

// newTree creates a new tree object from a C pointer. The function will set a finalizer for the object,
// thus no free is needed for it.
func (p *Parser) newTree(c uintptr) *Tree {
	base := &BaseTree{tls: p.tls, c: c}
	runtime.SetFinalizer(base, (*BaseTree).Close)

	newTree := &Tree{p: p, BaseTree: base, cache: make(map[C.TSNode]*Node)}
	return newTree
}

// Tree represents the syntax tree of an entire source code file
// Note: Tree instances are not thread safe;
// you must copy a tree if you want to use it on multiple threads simultaneously.
type Tree struct {
	*BaseTree

	// p is a pointer to a Parser that produced the Tree. Only used to keep Parser alive.
	// Otherwise Parser may be GC'ed (and deleted by the finalizer) while some Tree objects are still in use.
	p *Parser

	// most probably better save node.id
	cache map[C.TSNode]*Node
}

// Copy returns a new copy of a tree
func (t *Tree) Copy() *Tree {
	return t.p.newTree(C.Xts_tree_copy(t.p.tls, t.c))
}

// RootNode returns root node of a tree
func (t *Tree) RootNode() *Node {
	ptr := C.Xts_tree_root_node(t.p.tls, t.c)
	return t.cachedNode(ptr)
}

func (t *Tree) cachedNode(ptr C.TSNode) *Node {
	if ptr.Id == 0 {
		return nil
	}

	if n, ok := t.cache[ptr]; ok {
		return n
	}

	n := &Node{ptr, t}
	t.cache[ptr] = n
	return n
}

// Close should be called to ensure that all the memory used by the tree is freed.
func (t *BaseTree) Close() {
	if !t.isClosed {
		C.Xts_tree_delete(t.tls, t.c)
	}

	t.isClosed = true
}

type EditInput struct {
	StartIndex  uint32
	OldEndIndex uint32
	NewEndIndex uint32
	StartPoint  Point
	OldEndPoint Point
	NewEndPoint Point
}

func (i EditInput) c(tls *libc.TLS) uintptr {
	ptr := libc.Xmalloc(tls, types.Size_t(unsafe.Sizeof(C.TSInputEdit{})))
	c := (*C.TSInputEdit)(unsafe.Pointer(ptr))
	c.Start_byte = uint32(i.StartIndex)
	c.Old_end_byte = uint32(i.OldEndIndex)
	c.New_end_byte = uint32(i.NewEndIndex)
	c.Start_point = C.TSPoint{
		Row:    uint32(i.StartPoint.Row),
		Column: uint32(i.StartPoint.Column),
	}
	c.Old_end_point = C.TSPoint{
		Row:    uint32(i.OldEndPoint.Row),
		Column: uint32(i.OldEndPoint.Column),
	}
	c.New_end_point = C.TSPoint{
		Row:    uint32(i.OldEndPoint.Row),
		Column: uint32(i.OldEndPoint.Column),
	}
	return ptr
}

// Edit the syntax tree to keep it in sync with source code that has been edited.
func (t *Tree) Edit(i EditInput) {
	ic := i.c(t.p.tls)
	defer libc.Xfree(t.p.tls, ic)
	C.Xts_tree_edit(t.p.tls, t.c, ic)
}

// Language defines how to parse a particular programming language
type Language = lang.Language

// Node represents a single node in the syntax tree
// It tracks its start and end positions in the source code,
// as well as its relation to other nodes like its parent, siblings and children.
type Node struct {
	c C.TSNode
	t *Tree // keep pointer on tree because node is valid only as long as tree is
}

type Symbol = C.TSSymbol

type SymbolType = lang.SymbolType

const (
	SymbolTypeRegular SymbolType = iota
	SymbolTypeAnonymous
	SymbolTypeAuxiliary
)

// StartByte returns the node's start byte.
func (n Node) StartByte() uint32 {
	return uint32(C.Xts_node_start_byte(n.t.tls, n.c))
}

// EndByte returns the node's end byte.
func (n Node) EndByte() uint32 {
	return uint32(C.Xts_node_end_byte(n.t.tls, n.c))
}

// StartPoint returns the node's start position in terms of rows and columns.
func (n Node) StartPoint() Point {
	p := C.Xts_node_start_point(n.t.tls, n.c)
	return Point{
		Row:    uint32(p.Row),
		Column: uint32(p.Column),
	}
}

// EndPoint returns the node's end position in terms of rows and columns.
func (n Node) EndPoint() Point {
	p := C.Xts_node_end_point(n.t.tls, n.c)
	return Point{
		Row:    uint32(p.Row),
		Column: uint32(p.Column),
	}
}

// Symbol returns the node's type as a Symbol.
func (n Node) Symbol() Symbol {
	return C.Xts_node_symbol(n.t.tls, n.c)
}

// Type returns the node's type as a string.
func (n Node) Type() string {
	return libc.GoString(C.Xts_node_type(n.t.tls, n.c))
}

// String returns an S-expression representing the node as a string.
func (n Node) String() string {
	ptr := C.Xts_node_string(n.t.tls, n.c)
	defer libc.Xfree(n.t.tls, ptr)
	return libc.GoString(ptr)
}

// Equal checks if two nodes are identical.
func (n Node) Equal(other *Node) bool {
	return C.Xts_node_eq(n.t.tls, n.c, other.c) != 0
}

// IsNull checks if the node is null.
func (n Node) IsNull() bool {
	return C.Xts_node_is_null(n.t.tls, n.c) != 0
}

// IsNamed checks if the node is *named*.
// Named nodes correspond to named rules in the grammar,
// whereas *anonymous* nodes correspond to string literals in the grammar.
func (n Node) IsNamed() bool {
	return C.Xts_node_is_named(n.t.tls, n.c) != 0
}

// IsMissing checks if the node is *missing*.
// Missing nodes are inserted by the parser in order to recover from certain kinds of syntax errors.
func (n Node) IsMissing() bool {
	return C.Xts_node_is_missing(n.t.tls, n.c) != 0
}

// HasChanges checks if a syntax node has been edited.
func (n Node) HasChanges() bool {
	return C.Xts_node_has_changes(n.t.tls, n.c) != 0
}

// HasError check if the node is a syntax error or contains any syntax errors.
func (n Node) HasError() bool {
	return C.Xts_node_has_error(n.t.tls, n.c) != 0
}

// Parent returns the node's immediate parent.
func (n Node) Parent() *Node {
	nn := C.Xts_node_parent(n.t.tls, n.c)
	return n.t.cachedNode(nn)
}

// Child returns the node's child at the given index, where zero represents the first child.
func (n Node) Child(idx int) *Node {
	nn := C.Xts_node_child(n.t.tls, n.c, uint32(idx))
	return n.t.cachedNode(nn)
}

// NamedChild returns the node's *named* child at the given index.
func (n Node) NamedChild(idx int) *Node {
	nn := C.Xts_node_named_child(n.t.tls, n.c, uint32(idx))
	return n.t.cachedNode(nn)
}

// ChildCount returns the node's number of children.
func (n Node) ChildCount() uint32 {
	return uint32(C.Xts_node_child_count(n.t.tls, n.c))
}

// NamedChildCount returns the node's number of *named* children.
func (n Node) NamedChildCount() uint32 {
	return uint32(C.Xts_node_named_child_count(n.t.tls, n.c))
}

// ChildByFieldName returns the node's child with the given field name.
func (n Node) ChildByFieldName(name string) *Node {
	str, _ := libc.CString(name)
	defer libc.Xfree(n.t.tls, str)
	nn := C.Xts_node_child_by_field_name(n.t.tls, n.c, str, uint32(len(name)))
	return n.t.cachedNode(nn)
}

// NextSibling returns the node's next sibling.
func (n Node) NextSibling() *Node {
	nn := C.Xts_node_next_sibling(n.t.tls, n.c)
	return n.t.cachedNode(nn)
}

// NextNamedSibling returns the node's next *named* sibling.
func (n Node) NextNamedSibling() *Node {
	nn := C.Xts_node_next_named_sibling(n.t.tls, n.c)
	return n.t.cachedNode(nn)
}

// PrevSibling returns the node's previous sibling.
func (n Node) PrevSibling() *Node {
	nn := C.Xts_node_prev_sibling(n.t.tls, n.c)
	return n.t.cachedNode(nn)
}

// PrevNamedSibling returns the node's previous *named* sibling.
func (n Node) PrevNamedSibling() *Node {
	nn := C.Xts_node_prev_named_sibling(n.t.tls, n.c)
	return n.t.cachedNode(nn)
}

// Edit the node to keep it in-sync with source code that has been edited.
func (n *Node) Edit(i EditInput) {
	ic := i.c(n.t.tls)
	defer libc.Xfree(n.t.tls, ic)
	clonePtr := libc.Xmalloc(n.t.tls, types.Size_t(unsafe.Sizeof(C.TSNode{})))
	defer libc.Xfree(n.t.tls, clonePtr)
	clone := (*C.TSNode)(unsafe.Pointer(clonePtr))
	*clone = n.c
	C.Xts_node_edit(n.t.tls, clonePtr, ic)
	n.c = *clone
}

// Content returns node's source code from input as a string
func (n Node) Content(input []byte) string {
	return string(input[n.StartByte():n.EndByte()])
}

// TreeCursor allows you to walk a syntax tree more efficiently than is
// possible using the `Node` functions. It is a mutable object that is always
// on a certain syntax node, and can be moved imperatively to different nodes.
type TreeCursor struct {
	c uintptr
	t *Tree

	isClosed bool
}

// NewTreeCursor creates a new tree cursor starting from the given node.
func NewTreeCursor(n *Node) *TreeCursor {
	cc := C.Xts_tree_cursor_new(n.t.tls, n.c)
	c := &TreeCursor{
		c: libc.Xmalloc(n.t.tls, types.Size_t(unsafe.Sizeof(C.TSTreeCursor{}))),
		t: n.t,
	}
	*(*C.TSTreeCursor)(unsafe.Pointer(c.c)) = cc

	runtime.SetFinalizer(c, (*TreeCursor).Close)
	return c
}

// Close should be called to ensure that all the memory used by the tree cursor
// is freed.
func (c *TreeCursor) Close() {
	if !c.isClosed {
		C.Xts_tree_cursor_delete(c.t.tls, c.c)
		libc.Xfree(c.t.tls, c.c)
	}

	c.isClosed = true
}

// Reset re-initializes a tree cursor to start at a different node.
func (c *TreeCursor) Reset(n *Node) {
	c.t = n.t
	C.Xts_tree_cursor_reset(c.t.tls, c.c, n.c)
}

// CurrentNode of the tree cursor.
func (c *TreeCursor) CurrentNode() *Node {
	n := C.Xts_tree_cursor_current_node(c.t.tls, c.c)
	return c.t.cachedNode(n)
}

// CurrentFieldName gets the field name of the tree cursor's current node.
//
// This returns empty string if the current node doesn't have a field.
func (c *TreeCursor) CurrentFieldName() string {
	return libc.GoString(C.Xts_tree_cursor_current_field_name(c.t.tls, c.c))
}

// GoToParent moves the cursor to the parent of its current node.
//
// This returns `true` if the cursor successfully moved, and returns `false`
// if there was no parent node (the cursor was already on the root node).
func (c *TreeCursor) GoToParent() bool {
	return C.Xts_tree_cursor_goto_parent(c.t.tls, c.c) != 0
}

// GoToNextSibling moves the cursor to the next sibling of its current node.
//
// This returns `true` if the cursor successfully moved, and returns `false`
// if there was no next sibling node.
func (c *TreeCursor) GoToNextSibling() bool {
	return C.Xts_tree_cursor_goto_next_sibling(c.t.tls, c.c) != 0
}

// GoToFirstChild moves the cursor to the first child of its current node.
//
// This returns `true` if the cursor successfully moved, and returns `false`
// if there were no children.
func (c *TreeCursor) GoToFirstChild() bool {
	return C.Xts_tree_cursor_goto_first_child(c.t.tls, c.c) != 0
}

// GoToFirstChildForByte moves the cursor to the first child of its current node
// that extends beyond the given byte offset.
//
// This returns the index of the child node if one was found, and returns -1
// if no such child was found.
func (c *TreeCursor) GoToFirstChildForByte(b uint32) int64 {
	return C.Xts_tree_cursor_goto_first_child_for_byte(c.t.tls, c.c, uint32(b))
}

// QueryErrorType - value that indicates the type of QueryError.
type QueryErrorType int

const (
	QueryErrorNone QueryErrorType = iota
	QueryErrorSyntax
	QueryErrorNodeType
	QueryErrorField
	QueryErrorCapture
)

// QueryError - if there is an error in the query,
// then the Offset argument will be set to the byte offset of the error,
// and the Type argument will be set to a value that indicates the type of error.
type QueryError struct {
	Offset uint32
	Type   QueryErrorType
}

func (qe *QueryError) Error() string {
	switch qe.Type {
	case QueryErrorNone:
		return ""

	case QueryErrorSyntax:
		return fmt.Sprintf("syntax error (offset: %d)", qe.Offset)

	case QueryErrorNodeType:
		return fmt.Sprintf("node type error (offset: %d)", qe.Offset)

	case QueryErrorField:
		return fmt.Sprintf("field error (offset: %d)", qe.Offset)

	case QueryErrorCapture:
		return fmt.Sprintf("capture error (offset: %d)", qe.Offset)

	default:
		return fmt.Sprintf("unknown error (offset: %d)", qe.Offset)
	}
}

// Query API
type Query struct {
	tls      *libc.TLS
	c        uintptr
	isClosed bool
}

// NewQuery creates a query by specifying a string containing one or more patterns.
// In case of error returns QueryError.
func NewQuery(pattern []byte, l *Language) (*Query, error) {
	tls := libc.NewTLS()
	input := cbytes(tls, pattern)
	defer libc.Xfree(tls, input)
	erroff := libc.Xmalloc(tls, types.Size_t(unsafe.Sizeof(uint32(0))))
	defer libc.Xfree(tls, erroff)
	errtype := libc.Xmalloc(tls, types.Size_t(unsafe.Sizeof(C.TSQueryError(0))))
	defer libc.Xfree(tls, errtype)

	c := C.Xts_query_new(
		tls,
		lang.LanguagePtr(l),
		input,
		uint32(len(pattern)),
		erroff,
		errtype,
	)
	if errtype := *(*C.TSQueryError)(unsafe.Pointer(errtype)); errtype != C.TSQueryError(QueryErrorNone) {
		tls.Close()
		erroff := *(*uint32)(unsafe.Pointer(erroff))
		return nil, &QueryError{Offset: uint32(erroff), Type: QueryErrorType(errtype)}
	}

	q := &Query{tls: tls, c: c}
	runtime.SetFinalizer(q, (*Query).Close)

	return q, nil
}

// Close should be called to ensure that all the memory used by the query is freed.
func (q *Query) Close() {
	if !q.isClosed {
		C.Xts_query_delete(q.tls, q.c)
		q.tls.Close()
	}

	q.isClosed = true
}

func (q *Query) PatternCount() uint32 {
	return uint32(C.Xts_query_pattern_count(q.tls, q.c))
}

func (q *Query) CaptureCount() uint32 {
	return uint32(C.Xts_query_capture_count(q.tls, q.c))
}

func (q *Query) StringCount() uint32 {
	return uint32(C.Xts_query_string_count(q.tls, q.c))
}

type QueryPredicateStepType int

const (
	QueryPredicateStepTypeDone QueryPredicateStepType = iota
	QueryPredicateStepTypeCapture
	QueryPredicateStepTypeString
)

type QueryPredicateStep struct {
	Type    QueryPredicateStepType
	ValueId uint32
}

func (q *Query) PredicatesForPattern(patternIndex uint32) []QueryPredicateStep {
	lengthPtr := libc.Xmalloc(q.tls, types.Size_t(unsafe.Sizeof(uint32(0))))
	defer libc.Xfree(q.tls, lengthPtr)
	cPredicateStep := C.Xts_query_predicates_for_pattern(q.tls, q.c, uint32(patternIndex), lengthPtr)
	count := int(*(*uint32)(unsafe.Pointer(lengthPtr)))
	predicateSteps := make([]QueryPredicateStep, 0, count)

	for i := 0; i < count; i++ {
		s := (*C.TSQueryPredicateStep)(unsafe.Pointer(cPredicateStep + uintptr(i)*unsafe.Sizeof(C.TSQueryPredicateStep{})))
		predicateSteps = append(predicateSteps, QueryPredicateStep{
			Type:    QueryPredicateStepType(s.Type),
			ValueId: uint32(s.Value_id),
		})
	}

	return predicateSteps
}

func (q *Query) CaptureNameForId(id uint32) string {
	lengthPtr := libc.Xmalloc(q.tls, types.Size_t(unsafe.Sizeof(uint32(0))))
	defer libc.Xfree(q.tls, lengthPtr)
	name := C.Xts_query_capture_name_for_id(q.tls, q.c, uint32(id), lengthPtr)
	return goStringN(name, int(*(*uint32)(unsafe.Pointer(lengthPtr))))
}

func (q *Query) StringValueForId(id uint32) string {
	lengthPtr := libc.Xmalloc(q.tls, types.Size_t(unsafe.Sizeof(uint32(0))))
	defer libc.Xfree(q.tls, lengthPtr)
	value := C.Xts_query_string_value_for_id(q.tls, q.c, uint32(id), lengthPtr)
	return goStringN(value, int(*(*uint32)(unsafe.Pointer(lengthPtr))))
}

// QueryCursor carries the state needed for processing the queries.
type QueryCursor struct {
	tls *libc.TLS
	c   uintptr
	t   *Tree

	isClosed bool
}

// NewQueryCursor creates a query cursor.
func NewQueryCursor() *QueryCursor {
	tls := libc.NewTLS()
	qc := &QueryCursor{
		tls: tls,
		c:   C.Xts_query_cursor_new(tls),
		t:   nil,
	}
	runtime.SetFinalizer(qc, (*QueryCursor).Close)

	return qc
}

// Exec executes the query on a given syntax node.
func (qc *QueryCursor) Exec(q *Query, n *Node) {
	qc.t = n.t
	C.Xts_query_cursor_exec(qc.tls, qc.c, q.c, n.c)
}

func (qc *QueryCursor) SetPointRange(startPoint Point, endPoint Point) {
	cStartPoint := C.TSPoint{
		Row:    uint32(startPoint.Row),
		Column: uint32(startPoint.Column),
	}
	cEndPoint := C.TSPoint{
		Row:    uint32(endPoint.Row),
		Column: uint32(endPoint.Column),
	}
	C.Xts_query_cursor_set_point_range(qc.tls, qc.c, cStartPoint, cEndPoint)
}

// Close should be called to ensure that all the memory used by the query
// cursor is freed.
func (qc *QueryCursor) Close() {
	if !qc.isClosed {
		C.Xts_query_cursor_delete(qc.tls, qc.c)
		qc.tls.Close()
	}

	qc.isClosed = true
}

// QueryCapture is a captured node by a query with an index
type QueryCapture struct {
	Index uint32
	Node  *Node
}

// QueryMatch - you can then iterate over the matches.
type QueryMatch struct {
	ID           uint32
	PatternIndex uint16
	Captures     []QueryCapture
}

func (qc *QueryCursor) queryMatchFromC(cqmPtr uintptr) *QueryMatch {
	cqm := (*C.TSQueryMatch)(unsafe.Pointer(cqmPtr))
	count := int(cqm.Capture_count)
	qm := &QueryMatch{
		ID:           uint32(cqm.Id),
		PatternIndex: uint16(cqm.Pattern_index),
		Captures:     make([]QueryCapture, 0, count),
	}
	for i := 0; i < count; i++ {
		c := (*C.TSQueryCapture)(unsafe.Pointer(cqm.Captures + uintptr(i)*unsafe.Sizeof(C.TSQueryCapture{})))
		idx := uint32(c.Index)
		node := qc.t.cachedNode(c.Node)
		qm.Captures = append(qm.Captures, QueryCapture{idx, node})
	}
	return qm
}

// NextMatch iterates over matches.
// This function will return (nil, false) when there are no more matches.
// Otherwise, it will populate the QueryMatch with data
// about which pattern matched and which nodes were captured.
func (qc *QueryCursor) NextMatch() (*QueryMatch, bool) {
	cqmPtr := libc.Xmalloc(qc.tls, types.Size_t(unsafe.Sizeof(C.TSQueryMatch{})))
	defer libc.Xfree(qc.tls, cqmPtr)
	if C.Xts_query_cursor_next_match(qc.tls, qc.c, cqmPtr) == 0 {
		return nil, false
	}
	return qc.queryMatchFromC(cqmPtr), true
}

func (qc *QueryCursor) NextCapture() (*QueryMatch, uint32, bool) {
	cqmPtr := libc.Xmalloc(qc.tls, types.Size_t(unsafe.Sizeof(C.TSQueryMatch{})))
	defer libc.Xfree(qc.tls, cqmPtr)
	captureIndexPtr := libc.Xmalloc(qc.tls, types.Size_t(unsafe.Sizeof(uint32(0))))
	defer libc.Xfree(qc.tls, cqmPtr)
	if C.Xts_query_cursor_next_capture(qc.tls, qc.c, cqmPtr, captureIndexPtr) == 0 {
		return nil, 0, false
	}
	qm := qc.queryMatchFromC(cqmPtr)
	return qm, *(*uint32)(unsafe.Pointer(captureIndexPtr)), true
}

// keeps callbacks for parser.parse method
type readFuncsMap struct {
	sync.Mutex

	funcs map[int]ReadFunc
	count int
}

func (m *readFuncsMap) register(f ReadFunc) int {
	m.Lock()
	defer m.Unlock()

	m.count++
	m.funcs[m.count] = f
	return m.count
}

func (m *readFuncsMap) unregister(id int) {
	m.Lock()
	defer m.Unlock()

	delete(m.funcs, id)
}

func (m *readFuncsMap) get(id int) ReadFunc {
	m.Lock()
	defer m.Unlock()

	return m.funcs[id]
}

func cbytes(tls *libc.TLS, b []byte) uintptr {
	cb := libc.Xmalloc(tls, types.Size_t(len(b)))
	for i, bb := range b {
		*(*byte)(unsafe.Pointer(cb + uintptr(i))) = bb
	}
	return cb
}

func goStringN(s uintptr, n int) string {
	if s == 0 {
		return ""
	}
	var buf strings.Builder
	buf.Grow(n)
	for i := 0; i < n; i++ {
		buf.WriteByte(*(*byte)(unsafe.Pointer(s)))
		s++
	}
	return buf.String()
}
