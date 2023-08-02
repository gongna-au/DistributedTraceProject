package config

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/big"
	"time"
)

const (
	_            Type = iota
	TypeMaster        // force route to master node
	TypeSlave         // force route to slave node
	TypeRoute         // custom route
	TypeFullScan      // enable full-scan
	TypeDirect        // direct route
	TypeTrace         // distributed tracing
)

type Trace struct {
	Type    string `default:"jaeger" yaml:"type" json:"type"`
	Address string `default:"http://localhost:14268/api/traces" yaml:"address" json:"address"`
}

type Context struct {
	context.Context
	C FrontConn

	// sql Data
	Data []byte

	Stmt *Stmt
}

type Stmt struct {
	StatementID uint32
	PrepareStmt string
	ParamsCount uint16
	ParamsType  []int32
	ColumnNames []string
	BindVars    map[string]Value
	Hints       []*Hint
	StmtNode    StmtNode
}
type Hint struct {
	Type   Type
	Inputs []KeyValue
}

// KeyValue represents a pair of key and value.
type KeyValue struct {
	K string // key (optional)
	V string // value
}

// Node is the basic element of the AST.
// Interfaces embed Node should have 'Node' name suffix.
type Node interface {
	// Restore returns the sql text from ast tree
	Restore(ctx *RestoreCtx) error
	// Accept accepts Visitor to visit itself.
	// The returned node should replace original node.
	// ok returns false to stop visiting.
	//
	// Implementation of this method should first call visitor.Enter,
	// assign the returned node to its method receiver, if skipChildren returns true,
	// children should be skipped. Otherwise, call its children in particular order that
	// later elements depends on former elements. Finally, return visitor.Leave.
	Accept(v Visitor) (node Node, ok bool)
	// Text returns the utf8 encoding text of the element.
	Text() string
	// OriginalText returns the original text of the element.
	OriginalText() string
	// SetText sets original text to the Node.
	SetText(enc Encoding, text string)
	// SetOriginTextPosition set the start offset of this node in the origin text.
	SetOriginTextPosition(offset int)
	// OriginTextPosition get the start offset of this node in the origin text.
	OriginTextPosition() int
}

// Encoding provide encode/decode functions for a string with a specific charset.
type Encoding interface {
	// Name is the name of the encoding.
	Name() string
	// Tp is the type of the encoding.
	Tp() EncodingTp
	// Peek returns the next char.
	Peek(src []byte) []byte
	// MbLen returns multiple byte length, if the next character is single byte, return 0.
	MbLen(string) int
	// IsValid checks whether the utf-8 bytes can be convert to valid string in current encoding.
	IsValid(src []byte) bool
	// Foreach iterates the characters in in current encoding.
	Foreach(src []byte, op Op, fn func(from, to []byte, ok bool) bool)
	// Transform map the bytes in src to dest according to Op.
	// **the caller should initialize the dest if it wants to avoid memory alloc every time, or else it will always make a new one**
	// **the returned array may be the alias of `src`, edit the returned array on your own risk**
	Transform(dest *bytes.Buffer, src []byte, op Op) ([]byte, error)
	// ToUpper change a string to uppercase.
	ToUpper(src string) string
	// ToLower change a string to lowercase.
	ToLower(src string) string
}

// Op is used by Encoding.Transform.
type Op int16

type EncodingTp int8

type RestoreFlags uint64

type RestoreCtx struct {
	Flags     RestoreFlags
	In        io.Writer
	DefaultDB string
	CTENames  []string
}

type StmtNode interface {
	Node

	// Hints returns the arana hints.
	Hints() []string

	statement()
}

// Type represents the type of Hint.
type Type uint8

type FrontConn interface {
	// ID returns connection id.
	ID() uint32

	// Schema returns the current schema.
	Schema() string

	// SetSchema sets the current schema.
	SetSchema(schema string)

	// Tenant returns the tenant.
	Tenant() string

	// SetTenant sets the tenant.
	SetTenant(tenant string)

	// TransientVariables returns the transient variables.
	TransientVariables() map[string]Value

	// SetTransientVariables sets the transient variables.
	SetTransientVariables(v map[string]Value)

	// CharacterSet returns the character set.
	CharacterSet() uint8

	// ServerVersion returns the server version.
	ServerVersion() string
}

// Visitor visits a Node.
type Visitor interface {
	// Enter is called before children nodes are visited.
	// The returned node must be the same type as the input node n.
	// skipChildren returns true means children nodes should be skipped,
	// this is useful when work is done in Enter and there is no need to visit children.
	Enter(n Node) (node Node, skipChildren bool)
	// Leave is called after children nodes have been visited.
	// The returned node's type can be different from the input node if it is a ExprNode,
	// Non-expression node must be the same type as the input node n.
	// ok returns false to stop visiting.
	Leave(n Node) (node Node, ok bool)
}

type Value interface {
	fmt.Stringer
	Family() ValueFamily
	Float64() (float64, error)
	Int64() (int64, error)
	Uint64() (uint64, error)
	Decimal() (Decimal, error)
	Bool() (bool, error)
	Time() (time.Time, error)
	Less(than Value) bool
}

type ValueFamily uint8

type Decimal struct {
	value *big.Int

	// NOTE(vadim): this must be an int32, because we cast it to float64 during
	// calculations. If exp is 64 bit, we might lose precision.
	// If we cared about being able to represent every possible decimal, we
	// could make exp a *big.Int but it would hurt performance and numbers
	// like that are unrealistic.
	exp int32
}
