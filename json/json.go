package json

import (
	sitter "github.com/yourbase/treesitter"
	"github.com/yourbase/treesitter/internal/json"
	"github.com/yourbase/treesitter/internal/lang"
	"modernc.org/libc"
)

func GetLanguage() *sitter.Language {
	tls := libc.NewTLS()
	defer tls.Close()
	return lang.NewLanguage(json.Xtree_sitter_json(tls))
}
