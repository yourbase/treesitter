package python

import (
	sitter "github.com/yourbase/treesitter"
	"github.com/yourbase/treesitter/internal/lang"
	"github.com/yourbase/treesitter/internal/python"
	"modernc.org/libc"
)

func GetLanguage() *sitter.Language {
	tls := libc.NewTLS()
	defer tls.Close()
	return lang.NewLanguage(python.Xtree_sitter_python(tls))
}
