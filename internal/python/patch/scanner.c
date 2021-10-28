// Copied from https://github.com/tree-sitter/tree-sitter-python/blob/79e014734f40fd37644af24b49f368ed6c75a501/src/scanner.cc

#include <tree_sitter/parser.h>
#include <wctype.h>
#include <string.h>
#include <assert.h>
#include <stdlib.h>

enum TokenType {
  NEWLINE,
  INDENT,
  DEDENT,
  STRING_START,
  STRING_CONTENT,
  STRING_END,
};

enum DelimiterFlags {
  SingleQuote = 1 << 0,
  DoubleQuote = 1 << 1,
  BackQuote = 1 << 2,
  Raw = 1 << 3,
  Format = 1 << 4,
  Triple = 1 << 5,
  Bytes = 1 << 6,
};

typedef uint8_t Delimiter;

static bool delimiter_is_format(Delimiter flags) {
  return (flags & Format) != 0;
}

static bool delimiter_is_raw(Delimiter flags) {
  return (flags & Raw) != 0;
}

static bool delimiter_is_triple(Delimiter flags) {
  return (flags & Triple) != 0;
}

static bool delimiter_is_bytes(Delimiter flags) {
  return (flags & Bytes) != 0;
}

static int32_t delimiter_end_character(Delimiter flags) {
  if (flags & SingleQuote) return '\'';
  if (flags & DoubleQuote) return '"';
  if (flags & BackQuote) return '`';
  return 0;
}

static void delimiter_set_end_character(Delimiter* flags, int32_t character) {
  switch (character) {
    case '\'':
      *flags |= SingleQuote;
      break;
    case '"':
      *flags |= DoubleQuote;
      break;
    case '`':
      *flags |= BackQuote;
      break;
    default:
      assert(false);
  }
}

typedef struct Scanner {
  uint16_t* indent_length_stack;
  size_t indent_length_stack_len;
  size_t indent_length_stack_cap;

  Delimiter* delimiter_stack;
  size_t delimiter_stack_len;
  size_t delimiter_stack_cap;
} Scanner;

static void scanner_push_indent_length(Scanner* s, uint16_t n) {
  s->indent_length_stack_len++;
  if (s->indent_length_stack_len > s->indent_length_stack_cap) {
    s->indent_length_stack = realloc(s->indent_length_stack, sizeof(uint16_t) * s->indent_length_stack_len);
    s->indent_length_stack_cap = s->indent_length_stack_len;
  }
  s->indent_length_stack[s->indent_length_stack_len - 1] = n;
}

static void scanner_push_delimiter(Scanner* s, Delimiter d) {
  s->delimiter_stack_len++;
  if (s->delimiter_stack_len > s->delimiter_stack_cap) {
    s->delimiter_stack = realloc(s->delimiter_stack, sizeof(Delimiter) * s->delimiter_stack_len);
    s->delimiter_stack_cap = s->delimiter_stack_len;
  }
  s->delimiter_stack[s->delimiter_stack_len - 1] = d;
}

static unsigned scanner_serialize(Scanner* s, char *buffer) {
  size_t i = 0;

  size_t stack_size = s->delimiter_stack_len;
  if (stack_size > UINT8_MAX) stack_size = UINT8_MAX;
  buffer[i++] = stack_size;

  memcpy(&buffer[i], s->delimiter_stack, stack_size);
  i += stack_size;

  uint16_t* iter = s->indent_length_stack + 1;
  uint16_t* end = s->indent_length_stack + s->indent_length_stack_len;

  for (; iter != end && i < TREE_SITTER_SERIALIZATION_BUFFER_SIZE; ++iter) {
    buffer[i++] = *iter;
  }

  return i;
}

static void scanner_deserialize(Scanner* s, const char *buffer, unsigned length) {
  s->delimiter_stack_len = 0;
  s->indent_length_stack_len = 0;
  scanner_push_indent_length(s, 0);

  if (length > 0) {
    size_t i = 0;

    size_t delimiter_count = (uint8_t)buffer[i++];
    s->delimiter_stack = realloc(s->delimiter_stack, sizeof(Delimiter) * delimiter_count);
    s->delimiter_stack_len = delimiter_count;
    s->delimiter_stack_cap = delimiter_count;
    memcpy(s->delimiter_stack, &buffer[i], delimiter_count);
    i += delimiter_count;

    for (; i < length; i++) {
      scanner_push_indent_length(s, buffer[i]);
    }
  }
}

static Scanner* new_scanner() {
  assert(sizeof(Delimiter) == sizeof(char));
  Scanner* s = malloc(sizeof(Scanner));
  s->indent_length_stack = NULL;
  s->indent_length_stack_len = 0;
  s->indent_length_stack_cap = 0;
  s->delimiter_stack = NULL;
  s->delimiter_stack_len = 0;
  s->delimiter_stack_cap = 0;
  scanner_deserialize(s, NULL, 0);
  return s;
}

static bool scanner_scan(Scanner* s, TSLexer *lexer, const bool *valid_symbols) {
  if (valid_symbols[STRING_CONTENT] && !valid_symbols[INDENT] && s->delimiter_stack_len != 0) {
    Delimiter delimiter = s->delimiter_stack[s->delimiter_stack_len - 1];
    int32_t end_character = delimiter_end_character(delimiter);
    bool has_content = false;
    while (lexer->lookahead) {
      if (lexer->lookahead == '{' && delimiter_is_format(delimiter)) {
        lexer->mark_end(lexer);
        lexer->advance(lexer, false);
        if (lexer->lookahead == '{') {
          lexer->advance(lexer, false);
        } else {
          lexer->result_symbol = STRING_CONTENT;
          return has_content;
        }
      } else if (lexer->lookahead == '\\') {
        if (delimiter_is_raw(delimiter)) {
          lexer->advance(lexer, false);
        } else if (delimiter_is_bytes(delimiter)) {
            lexer->mark_end(lexer);
            lexer->advance(lexer, false);
            if (lexer->lookahead == 'N' || lexer->lookahead == 'u' || lexer->lookahead == 'U') {
              // In bytes string, \N{...}, \uXXXX and \UXXXXXXXX are not escape sequences
              // https://docs.python.org/3/reference/lexical_analysis.html#string-and-bytes-literals
              lexer->advance(lexer, false);
            } else {
                lexer->result_symbol = STRING_CONTENT;
                return has_content;
            }
        } else {
          lexer->mark_end(lexer);
          lexer->result_symbol = STRING_CONTENT;
          return has_content;
        }
      } else if (lexer->lookahead == end_character) {
        if (delimiter_is_triple(delimiter)) {
          lexer->mark_end(lexer);
          lexer->advance(lexer, false);
          if (lexer->lookahead == end_character) {
            lexer->advance(lexer, false);
            if (lexer->lookahead == end_character) {
              if (has_content) {
                lexer->result_symbol = STRING_CONTENT;
              } else {
                lexer->advance(lexer, false);
                lexer->mark_end(lexer);
                s->delimiter_stack_len--;
                lexer->result_symbol = STRING_END;
              }
              return true;
            }
          }
        } else {
          if (has_content) {
            lexer->result_symbol = STRING_CONTENT;
          } else {
            lexer->advance(lexer, false);
            s->delimiter_stack_len--;
            lexer->result_symbol = STRING_END;
          }
          lexer->mark_end(lexer);
          return true;
        }
      } else if (lexer->lookahead == '\n' && has_content && !delimiter_is_triple(delimiter)) {
        return false;
      }
      lexer->advance(lexer, false);
      has_content = true;
    }
  }

  lexer->mark_end(lexer);

  bool found_end_of_line = false;
  uint32_t indent_length = 0;
  int32_t first_comment_indent_length = -1;
  for (;;) {
    if (lexer->lookahead == '\n') {
      found_end_of_line = true;
      indent_length = 0;
      lexer->advance(lexer, true);
    } else if (lexer->lookahead == ' ') {
      indent_length++;
      lexer->advance(lexer, true);
    } else if (lexer->lookahead == '\r') {
      indent_length = 0;
      lexer->advance(lexer, true);
    } else if (lexer->lookahead == '\t') {
      indent_length += 8;
      lexer->advance(lexer, true);
    } else if (lexer->lookahead == '#') {
      if (first_comment_indent_length == -1) {
        first_comment_indent_length = (int32_t)indent_length;
      }
      while (lexer->lookahead && lexer->lookahead != '\n') {
        lexer->advance(lexer, true);
      }
      lexer->advance(lexer, true);
      indent_length = 0;
    } else if (lexer->lookahead == '\\') {
      lexer->advance(lexer, true);
      if (iswspace(lexer->lookahead)) {
        lexer->advance(lexer, true);
      } else {
        return false;
      }
    } else if (lexer->lookahead == '\f') {
      indent_length = 0;
      lexer->advance(lexer, true);
    } else if (lexer->lookahead == 0) {
      indent_length = 0;
      found_end_of_line = true;
      break;
    } else {
      break;
    }
  }

  if (found_end_of_line) {
    if (s->indent_length_stack_len != 0) {
      uint16_t current_indent_length = s->indent_length_stack[s->indent_length_stack_len - 1];

      if (
        valid_symbols[INDENT] &&
        indent_length > current_indent_length
      ) {
        scanner_push_indent_length(s, indent_length);
        lexer->result_symbol = INDENT;
        return true;
      }

      if (
        valid_symbols[DEDENT] &&
        indent_length < current_indent_length &&

        // Wait to create a dedent token until we've consumed any comments
        // whose indentation matches the current block.
        first_comment_indent_length < (int32_t)current_indent_length
      ) {
        s->indent_length_stack_len--;
        lexer->result_symbol = DEDENT;
        return true;
      }
    }

    if (valid_symbols[NEWLINE]) {
      lexer->result_symbol = NEWLINE;
      return true;
    }
  }

  if (first_comment_indent_length == -1 && valid_symbols[STRING_START]) {
    Delimiter delimiter;

    bool has_flags = false;
    while (lexer->lookahead) {
      if (lexer->lookahead == 'f' || lexer->lookahead == 'F') {
        delimiter |= Format;
      } else if (lexer->lookahead == 'r' || lexer->lookahead == 'R') {
        delimiter |= Raw;
      } else if (lexer->lookahead == 'b' || lexer->lookahead == 'B') {
        delimiter |= Bytes;
      } else if (lexer->lookahead != 'u' && lexer->lookahead != 'U') {
        break;
      }
      has_flags = true;
      lexer->advance(lexer, false);
    }

    if (lexer->lookahead == '`') {
      delimiter_set_end_character(&delimiter, '`');
      lexer->advance(lexer, false);
      lexer->mark_end(lexer);
    } else if (lexer->lookahead == '\'') {
      delimiter_set_end_character(&delimiter, '\'');
      lexer->advance(lexer, false);
      lexer->mark_end(lexer);
      if (lexer->lookahead == '\'') {
        lexer->advance(lexer, false);
        if (lexer->lookahead == '\'') {
          lexer->advance(lexer, false);
          lexer->mark_end(lexer);
          delimiter |= Triple;
        }
      }
    } else if (lexer->lookahead == '"') {
      delimiter_set_end_character(&delimiter, '"');
      lexer->advance(lexer, false);
      lexer->mark_end(lexer);
      if (lexer->lookahead == '"') {
        lexer->advance(lexer, false);
        if (lexer->lookahead == '"') {
          lexer->advance(lexer, false);
          lexer->mark_end(lexer);
          delimiter |= Triple;
        }
      }
    }

    if (delimiter_end_character(delimiter)) {
      scanner_push_delimiter(s, delimiter);
      lexer->result_symbol = STRING_START;
      return true;
    } else if (has_flags) {
      return false;
    }
  }

  return false;
}

// Exports

void *tree_sitter_python_external_scanner_create() {
  return new_scanner();
}

bool tree_sitter_python_external_scanner_scan(void *payload, TSLexer *lexer,
                                            const bool *valid_symbols) {
  Scanner *scanner = (Scanner*)(payload);
  return scanner_scan(scanner, lexer, valid_symbols);
}

unsigned tree_sitter_python_external_scanner_serialize(void *payload, char *buffer) {
  Scanner *scanner = (Scanner*)(payload);
  return scanner_serialize(scanner, buffer);
}

void tree_sitter_python_external_scanner_deserialize(void *payload, const char *buffer, unsigned length) {
  Scanner *scanner = (Scanner*)(payload);
  scanner_deserialize(scanner, buffer, length);
}

void tree_sitter_python_external_scanner_destroy(void *payload) {
  Scanner *scanner = (Scanner*)payload;
  if (scanner->delimiter_stack != NULL) {
    free(scanner->delimiter_stack);
  }
  if (scanner->indent_length_stack != NULL) {
    free(scanner->indent_length_stack);
  }
  free(scanner);
}
