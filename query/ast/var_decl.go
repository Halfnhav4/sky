package ast

import (
	"fmt"
	"github.com/skydb/sky/core"
)

// VarDecl represents a variable declaration in the query. The value
// of the variable persist for the duration of an object and can be
// referenced like any other property on the database. The variable can
// also be associated with another variable for the purpose of reusing
// factorization.
type VarDecl struct {
	Id          int64
	Name        string
	DataType    string
	Association string
	index       int
}

func (v *VarDecl) node() {}
func (v *VarDecl) statement() {}

// NewVarDecl returns a new VarDecl instance.
func NewVarDecl(id int64, name string, dataType string) *VarDecl {
	return &VarDecl{Id: id, Name: name, DataType: dataType}
}

// Returns the index of the variable declaration. This is used internally
// to track the generated struct position.
func (v *VarDecl) Index() int {
	return v.index
}

// Determines if the variable declaration is a permanent variable.
func (v *VarDecl) IsPermanent() bool {
	return (v.Id > 0)
}

// Determines if the variable declaration is a transient variable.
func (v *VarDecl) IsTransient() bool {
	return v.Id < 0
}

// Determines if the variable declaration is a system variable.
func (v *VarDecl) IsSystem() bool {
	return v.Id == 0 && (len(v.Name) != 0 && v.Name[0] == '@')
}

// Determines if the variable declaration is a declared variable.
func (v *VarDecl) IsDeclared() bool {
	return v.Id == 0 && !v.IsSystem()
}

func (v *VarDecl) String() string {
	var dataType string
	switch v.DataType {
	case core.FactorDataType:
		dataType = "FACTOR"
	case core.StringDataType:
		dataType = "STRING"
	case core.IntegerDataType:
		dataType = "INTEGER"
	case core.FloatDataType:
		dataType = "FLOAT"
	case core.BooleanDataType:
		dataType = "BOOLEAN"
	}

	str := fmt.Sprintf("DECLARE @%s AS %s", v.Name, dataType)
	if v.Association != "" {
		str += "(@" + v.Association + ")"
	}
	return str
}
