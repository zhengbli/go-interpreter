package object

import "fmt"

const (
	INTEGER_OBJ     = "INTEGER"
	BOOLEAN_OBJ     = "BOOLEAN"
	NULL_OBJ        = "NULL"
	RETURNVALUE_OBJ = "RETURNVALUE"
	ERROR_OBJ       = "ERROR"
)

type ObjectType string

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%v", b.Value) }

type Null struct {
}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "Null" }

type ReturnValue struct {
	Value Object
}

func (n *ReturnValue) Type() ObjectType { return RETURNVALUE_OBJ }
func (n *ReturnValue) Inspect() string  { return n.Value.Inspect() }

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return e.Message }
