// Code generated by 'ccgo -pkgname=lib -export-defines  -export-enums  -export-externs X -export-structs S -export-fields  -export-typedefs  -trace-translation-units -o internal/lib/treesitter_linux_arm.go -I upstream/tree-sitter/lib/src -I upstream/tree-sitter/lib/include upstream/tree-sitter/lib/src/get_changed_ranges.c upstream/tree-sitter/lib/src/language.c upstream/tree-sitter/lib/src/lexer.c upstream/tree-sitter/lib/src/lib.c upstream/tree-sitter/lib/src/node.c upstream/tree-sitter/lib/src/parser.c upstream/tree-sitter/lib/src/query.c upstream/tree-sitter/lib/src/stack.c upstream/tree-sitter/lib/src/subtree.c upstream/tree-sitter/lib/src/tree.c upstream/tree-sitter/lib/src/tree_cursor.c', DO NOT EDIT.

package lib

var CAPI = map[string]struct{}{
	"ts_external_scanner_state_copy":            {},
	"ts_external_scanner_state_data":            {},
	"ts_external_scanner_state_delete":          {},
	"ts_external_scanner_state_eq":              {},
	"ts_external_scanner_state_init":            {},
	"ts_language_field_count":                   {},
	"ts_language_field_id_for_name":             {},
	"ts_language_field_name_for_id":             {},
	"ts_language_public_symbol":                 {},
	"ts_language_symbol_count":                  {},
	"ts_language_symbol_for_name":               {},
	"ts_language_symbol_metadata":               {},
	"ts_language_symbol_name":                   {},
	"ts_language_symbol_type":                   {},
	"ts_language_table_entry":                   {},
	"ts_language_version":                       {},
	"ts_lexer_advance_to_end":                   {},
	"ts_lexer_delete":                           {},
	"ts_lexer_finish":                           {},
	"ts_lexer_included_ranges":                  {},
	"ts_lexer_init":                             {},
	"ts_lexer_mark_end":                         {},
	"ts_lexer_reset":                            {},
	"ts_lexer_set_included_ranges":              {},
	"ts_lexer_set_input":                        {},
	"ts_lexer_start":                            {},
	"ts_node_child":                             {},
	"ts_node_child_by_field_id":                 {},
	"ts_node_child_by_field_name":               {},
	"ts_node_child_count":                       {},
	"ts_node_descendant_for_byte_range":         {},
	"ts_node_descendant_for_point_range":        {},
	"ts_node_edit":                              {},
	"ts_node_end_byte":                          {},
	"ts_node_end_point":                         {},
	"ts_node_eq":                                {},
	"ts_node_field_name_for_child":              {},
	"ts_node_first_child_for_byte":              {},
	"ts_node_first_named_child_for_byte":        {},
	"ts_node_has_changes":                       {},
	"ts_node_has_error":                         {},
	"ts_node_is_extra":                          {},
	"ts_node_is_missing":                        {},
	"ts_node_is_named":                          {},
	"ts_node_is_null":                           {},
	"ts_node_named_child":                       {},
	"ts_node_named_child_count":                 {},
	"ts_node_named_descendant_for_byte_range":   {},
	"ts_node_named_descendant_for_point_range":  {},
	"ts_node_new":                               {},
	"ts_node_next_named_sibling":                {},
	"ts_node_next_sibling":                      {},
	"ts_node_parent":                            {},
	"ts_node_prev_named_sibling":                {},
	"ts_node_prev_sibling":                      {},
	"ts_node_start_byte":                        {},
	"ts_node_start_point":                       {},
	"ts_node_string":                            {},
	"ts_node_symbol":                            {},
	"ts_node_type":                              {},
	"ts_parser_cancellation_flag":               {},
	"ts_parser_delete":                          {},
	"ts_parser_included_ranges":                 {},
	"ts_parser_language":                        {},
	"ts_parser_logger":                          {},
	"ts_parser_new":                             {},
	"ts_parser_parse":                           {},
	"ts_parser_parse_string":                    {},
	"ts_parser_parse_string_encoding":           {},
	"ts_parser_print_dot_graphs":                {},
	"ts_parser_reset":                           {},
	"ts_parser_set_cancellation_flag":           {},
	"ts_parser_set_included_ranges":             {},
	"ts_parser_set_language":                    {},
	"ts_parser_set_logger":                      {},
	"ts_parser_set_timeout_micros":              {},
	"ts_parser_timeout_micros":                  {},
	"ts_query_capture_count":                    {},
	"ts_query_capture_name_for_id":              {},
	"ts_query_cursor__compare_captures":         {},
	"ts_query_cursor__compare_nodes":            {},
	"ts_query_cursor_delete":                    {},
	"ts_query_cursor_did_exceed_match_limit":    {},
	"ts_query_cursor_exec":                      {},
	"ts_query_cursor_match_limit":               {},
	"ts_query_cursor_new":                       {},
	"ts_query_cursor_next_capture":              {},
	"ts_query_cursor_next_match":                {},
	"ts_query_cursor_remove_match":              {},
	"ts_query_cursor_set_byte_range":            {},
	"ts_query_cursor_set_match_limit":           {},
	"ts_query_cursor_set_point_range":           {},
	"ts_query_delete":                           {},
	"ts_query_disable_capture":                  {},
	"ts_query_disable_pattern":                  {},
	"ts_query_new":                              {},
	"ts_query_pattern_count":                    {},
	"ts_query_predicates_for_pattern":           {},
	"ts_query_start_byte_for_pattern":           {},
	"ts_query_step_is_definite":                 {},
	"ts_query_string_count":                     {},
	"ts_query_string_value_for_id":              {},
	"ts_range_array_get_changed_ranges":         {},
	"ts_range_array_intersects":                 {},
	"ts_stack_can_merge":                        {},
	"ts_stack_clear":                            {},
	"ts_stack_copy_version":                     {},
	"ts_stack_delete":                           {},
	"ts_stack_dynamic_precedence":               {},
	"ts_stack_error_cost":                       {},
	"ts_stack_get_summary":                      {},
	"ts_stack_halt":                             {},
	"ts_stack_has_advanced_since_error":         {},
	"ts_stack_is_active":                        {},
	"ts_stack_is_halted":                        {},
	"ts_stack_is_paused":                        {},
	"ts_stack_iterate":                          {},
	"ts_stack_last_external_token":              {},
	"ts_stack_merge":                            {},
	"ts_stack_new":                              {},
	"ts_stack_node_count_since_error":           {},
	"ts_stack_pause":                            {},
	"ts_stack_pop_all":                          {},
	"ts_stack_pop_count":                        {},
	"ts_stack_pop_error":                        {},
	"ts_stack_pop_pending":                      {},
	"ts_stack_position":                         {},
	"ts_stack_print_dot_graph":                  {},
	"ts_stack_push":                             {},
	"ts_stack_record_summary":                   {},
	"ts_stack_remove_version":                   {},
	"ts_stack_renumber_version":                 {},
	"ts_stack_resume":                           {},
	"ts_stack_set_last_external_token":          {},
	"ts_stack_state":                            {},
	"ts_stack_swap_versions":                    {},
	"ts_stack_version_count":                    {},
	"ts_subtree__print_dot_graph":               {},
	"ts_subtree_array_clear":                    {},
	"ts_subtree_array_copy":                     {},
	"ts_subtree_array_delete":                   {},
	"ts_subtree_array_remove_trailing_extras":   {},
	"ts_subtree_array_reverse":                  {},
	"ts_subtree_balance":                        {},
	"ts_subtree_clone":                          {},
	"ts_subtree_compare":                        {},
	"ts_subtree_edit":                           {},
	"ts_subtree_eq":                             {},
	"ts_subtree_external_scanner_state_eq":      {},
	"ts_subtree_get_changed_ranges":             {},
	"ts_subtree_last_external_token":            {},
	"ts_subtree_make_mut":                       {},
	"ts_subtree_new_error":                      {},
	"ts_subtree_new_error_node":                 {},
	"ts_subtree_new_leaf":                       {},
	"ts_subtree_new_missing_leaf":               {},
	"ts_subtree_new_node":                       {},
	"ts_subtree_pool_delete":                    {},
	"ts_subtree_pool_new":                       {},
	"ts_subtree_print_dot_graph":                {},
	"ts_subtree_release":                        {},
	"ts_subtree_retain":                         {},
	"ts_subtree_set_symbol":                     {},
	"ts_subtree_string":                         {},
	"ts_subtree_summarize_children":             {},
	"ts_tree_copy":                              {},
	"ts_tree_cursor_copy":                       {},
	"ts_tree_cursor_current_field_id":           {},
	"ts_tree_cursor_current_field_name":         {},
	"ts_tree_cursor_current_node":               {},
	"ts_tree_cursor_current_status":             {},
	"ts_tree_cursor_delete":                     {},
	"ts_tree_cursor_goto_first_child":           {},
	"ts_tree_cursor_goto_first_child_for_byte":  {},
	"ts_tree_cursor_goto_first_child_for_point": {},
	"ts_tree_cursor_goto_next_sibling":          {},
	"ts_tree_cursor_goto_parent":                {},
	"ts_tree_cursor_init":                       {},
	"ts_tree_cursor_new":                        {},
	"ts_tree_cursor_parent_node":                {},
	"ts_tree_cursor_reset":                      {},
	"ts_tree_delete":                            {},
	"ts_tree_edit":                              {},
	"ts_tree_get_changed_ranges":                {},
	"ts_tree_language":                          {},
	"ts_tree_new":                               {},
	"ts_tree_print_dot_graph":                   {},
	"ts_tree_root_node":                         {},
}
