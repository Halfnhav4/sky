#include <stdlib.h>
#include "dbg.h"

#include "node.h"

//==============================================================================
//
// Functions
//
//==============================================================================

//--------------------------------------
// Lifecycle
//--------------------------------------

// Creates an AST node for a function return.
//
// value - The value to be returned.
// ret   - A pointer to where the ast node will be returned.
//
// Returns a function return node.
eql_ast_node *eql_ast_freturn_create(struct eql_ast_node *value)
{
    eql_ast_node *node = malloc(sizeof(eql_ast_node)); check_mem(node);
    node->type = EQL_AST_TYPE_FRETURN;
    node->parent = NULL;
    node->line_no = node->char_no = 0;
    node->generated = false;
    node->freturn.value = value;
    if(value) {
        value->parent = node;
    }

    return node;

error:
    eql_ast_node_free(node);
    return NULL;
}

// Frees a variable declaration AST node from memory.
//
// node - The AST node to free.
void eql_ast_freturn_free(struct eql_ast_node *node)
{
    if(node->freturn.value) eql_ast_node_free(node->freturn.value);
    node->freturn.value = NULL;
}

// Copies a node and its children.
//
// node - The node to copy.
// ret  - A pointer to where the new copy should be returned to.
//
// Returns 0 if successful, otherwise returns -1.
int eql_ast_freturn_copy(eql_ast_node *node, eql_ast_node **ret)
{
    int rc;
    check(node != NULL, "Node required");
    check(ret != NULL, "Return pointer required");

    eql_ast_node *clone = eql_ast_freturn_create(NULL);
    check_mem(clone);

    rc = eql_ast_node_copy(node->freturn.value, &clone->freturn.value);
    check(rc == 0, "Unable to copy return value");
    if(clone->freturn.value) clone->freturn.value->parent = clone;
    
    *ret = clone;
    return 0;

error:
    eql_ast_node_free(clone);
    *ret = NULL;
    return -1;
}


//--------------------------------------
// Codegen
//--------------------------------------

int eql_ast_freturn_codegen(eql_ast_node *node, eql_module *module,
                            LLVMValueRef *value)
{
    check(node != NULL, "Node is required");
    check(node->type == EQL_AST_TYPE_FRETURN, "Node must be a function return");
    
    LLVMBuilderRef builder = module->compiler->llvm_builder;

    // Return value if specified.
    if(node->freturn.value) {
        // Load return value.
        LLVMValueRef return_value = NULL;
        int rc = eql_ast_node_codegen(node->freturn.value, module, &return_value);
        check(rc == 0, "Unable to codegen function return value");
        check(return_value != NULL, "Missing return value");
        
        // Generate destroy for variable declarations.
        rc = eql_ast_block_codegen_destroy(node->parent, module);
        check(rc == 0, "Unable to generate block destroy");
        
        // Execute return of value.
        *value = LLVMBuildRet(builder, return_value);
        check(*value != NULL, "Unable to generate function return");
    }
    // Otherwise return void.
    else {
        *value = LLVMBuildRetVoid(builder);
        check(*value != NULL, "Unable to generate function return void");
    }
    
    return 0;

error:
    *value = NULL;
    return -1;
}


//--------------------------------------
// Preprocessor
//--------------------------------------

// Preprocess the node.
//
// node   - The node to validate.
// module - The module that the node is a part of.
//
// Returns 0 if successful, otherwise returns -1.
int eql_ast_freturn_preprocess(eql_ast_node *node, eql_module *module)
{
    int rc;
    check(node != NULL, "Node required");
    check(module != NULL, "Module required");

    // Preprocess value.
    if(node->freturn.value != NULL) {
        rc = eql_ast_node_preprocess(node->freturn.value, module);
        check(rc == 0, "Unable to preprocess return value");
    }

    return 0;

error:
    return -1;   
}


//--------------------------------------
// Validation
//--------------------------------------

// Validates the AST node.
//
// node   - The node to validate.
// module - The module that the node is a part of.
//
// Returns 0 if successful, otherwise returns -1.
int eql_ast_freturn_validate(eql_ast_node *node, eql_module *module)
{
    int rc;
    check(node != NULL, "Node required");
    check(module != NULL, "Module required");

    // Validate value.
    if(node->freturn.value != NULL) {
        rc = eql_ast_node_validate(node->freturn.value, module);
        check(rc == 0, "Unable to validate return value");
    }

    return 0;

error:
    return -1;   
}


//--------------------------------------
// Debugging
//--------------------------------------

// Append the contents of the AST node to the string.
// 
// node - The node to dump.
// ret  - A pointer to the bstring to concatenate to.
//
// Return 0 if successful, otherwise returns -1.s
int eql_ast_freturn_dump(eql_ast_node *node, bstring ret)
{
    int rc;
    check(node != NULL, "Node required");
    check(ret != NULL, "String required");

    // Append dump.
    check(bcatcstr(ret, "<freturn>\n") == BSTR_OK, "Unable to append dump");

    // Recursively dump children.
    if(node->freturn.value != NULL) {
        rc = eql_ast_node_dump(node->freturn.value, ret);
        check(rc == 0, "Unable to dump return value");
    }

    return 0;

error:
    return -1;
}