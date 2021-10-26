package lang

import (
	C "github.com/yourbase/treesitter/internal/lib"
	"modernc.org/libc"
)

// Language defines how to parse a particular programming language
type Language struct {
	ptr uintptr
}

// NewLanguage creates new Language from c pointer
func NewLanguage(ptr uintptr) *Language {
	return &Language{ptr}
}

func LanguagePtr(lang *Language) uintptr {
	return lang.ptr
}

// SymbolName returns a node type string for the given Symbol.
func (l *Language) SymbolName(s C.TSSymbol) string {
	tls := libc.NewTLS()
	defer tls.Close()
	return libc.GoString(C.Xts_language_symbol_name(tls, l.ptr, s))
}

// SymbolType returns named, anonymous, or a hidden type for a Symbol.
func (l *Language) SymbolType(s C.TSSymbol) SymbolType {
	tls := libc.NewTLS()
	defer tls.Close()
	return SymbolType(C.Xts_language_symbol_type(tls, l.ptr, s))
}

// SymbolCount returns the number of distinct field names in the language.
func (l *Language) SymbolCount() uint32 {
	tls := libc.NewTLS()
	defer tls.Close()
	return uint32(C.Xts_language_symbol_count(tls, l.ptr))
}

func (l *Language) FieldName(idx int) string {
	tls := libc.NewTLS()
	defer tls.Close()
	return libc.GoString(C.Xts_language_field_name_for_id(tls, l.ptr, uint16(idx)))
}

type SymbolType int

var symbolTypeNames = []string{
	"Regular",
	"Anonymous",
	"Auxiliary",
}

func (t SymbolType) String() string {
	return symbolTypeNames[t]
}
