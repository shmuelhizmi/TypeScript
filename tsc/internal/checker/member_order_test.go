package checker

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"gotest.tools/v3/assert"
)

// memberOrderFixture holds the container and the declarations of a generic interface. Each call of
// members builds the table of one instantiation: fresh symbols that share the declaration nodes.
type memberOrderFixture struct {
	container *ast.Symbol
	declared  *ast.Node
	inherited *ast.Node
}

func newMemberOrderFixture() memberOrderFixture {
	return memberOrderFixture{
		container: &ast.Symbol{Flags: ast.SymbolFlagsInterface, Declarations: []*ast.Node{{Loc: core.NewTextRange(100, 200)}}},
		declared:  &ast.Node{Loc: core.NewTextRange(120, 130)},
		inherited: &ast.Node{Loc: core.NewTextRange(20, 30)},
	}
}

func (f memberOrderFixture) members() ast.SymbolTable {
	return ast.SymbolTable{
		"zDeclared":  {Name: "zDeclared", Flags: ast.SymbolFlagsProperty, Declarations: []*ast.Node{f.declared}, ValueDeclaration: f.declared},
		"aInherited": {Name: "aInherited", Flags: ast.SymbolFlagsProperty, Declarations: []*ast.Node{f.inherited}, ValueDeclaration: f.inherited},
	}
}

func TestGetNamedMembersWithOrder(t *testing.T) {
	t.Parallel()
	c := &Checker{}
	c.compareSymbols = c.compareSymbolsWorker
	f := newMemberOrderFixture()
	want := []string{"zDeclared", "aInherited"}

	// The first instantiation records the order; the next one is projected through the recording.
	source := &InterfaceType{}
	assert.DeepEqual(t, symbolNames(c.getNamedMembersWithOrder(f.members(), f.container, source)), want)
	recorded := source.instantiatedMemberOrder
	assert.Assert(t, recorded != nil)
	second := f.members()
	properties := c.getNamedMembersWithOrder(second, f.container, source)
	assert.DeepEqual(t, symbolNames(properties), want)
	assert.Equal(t, properties[0], second["zDeclared"])
	assert.Equal(t, properties[1], second["aInherited"])
	assert.Equal(t, source.instantiatedMemberOrder, recorded)

	// Members that no longer match the recording get a fresh order, and the recording stays.
	changed := f.members()
	changed["aInherited"].ValueDeclaration = f.declared
	assert.DeepEqual(t, symbolNames(c.getNamedMembersWithOrder(changed, f.container, source)), []string{"aInherited", "zDeclared"})
	assert.Equal(t, source.instantiatedMemberOrder, recorded)

	// A single member is not worth recording.
	single := &InterfaceType{}
	members := f.members()
	delete(members, "aInherited")
	assert.DeepEqual(t, symbolNames(c.getNamedMembersWithOrder(members, f.container, single)), []string{"zDeclared"})
	assert.Assert(t, single.instantiatedMemberOrder == nil)

	// A container whose declarations changed after the recording gets a fresh order as well.
	f.container.Declarations = append(f.container.Declarations, &ast.Node{Loc: core.NewTextRange(10, 40)})
	assert.DeepEqual(t, symbolNames(c.getNamedMembersWithOrder(f.members(), f.container, source)), []string{"aInherited", "zDeclared"})
	assert.Equal(t, source.instantiatedMemberOrder, recorded)
}

func TestProjectMemberOrder(t *testing.T) {
	t.Parallel()
	c := &Checker{}
	c.compareSymbols = c.compareSymbolsWorker
	f := newMemberOrderFixture()
	members := f.members()
	order := createMemberOrder(members, c.getNamedMembers(members, f.container))
	assert.Equal(t, len(order), 2)
	assert.DeepEqual(t, symbolNames(projectMemberOrder(f.members(), order)), []string{"zDeclared", "aInherited"})

	mismatches := map[string]func(ast.SymbolTable){
		"missing member":             func(m ast.SymbolTable) { delete(m, "aInherited") },
		"extra member":               func(m ast.SymbolTable) { m["extra"] = &ast.Symbol{Name: "extra", Flags: ast.SymbolFlagsProperty} },
		"member under another name":  func(m ast.SymbolTable) { m["other"] = m["aInherited"]; delete(m, "aInherited") },
		"other first declaration":    func(m ast.SymbolTable) { m["zDeclared"].Declarations = []*ast.Node{f.inherited, f.declared} },
		"other value declaration":    func(m ast.SymbolTable) { m["zDeclared"].ValueDeclaration = f.inherited },
		"member that is not a value": func(m ast.SymbolTable) { m["zDeclared"].Flags = ast.SymbolFlagsInterface },
	}
	for name, mismatch := range mismatches {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			mismatched := f.members()
			mismatch(mismatched)
			assert.Assert(t, projectMemberOrder(mismatched, order) == nil)
		})
	}
}

func TestCreateMemberOrder(t *testing.T) {
	t.Parallel()
	c := &Checker{}
	c.compareSymbols = c.compareSymbolsWorker
	f := newMemberOrderFixture()
	members := f.members()
	properties := c.getNamedMembers(members, f.container)
	assert.Equal(t, len(createMemberOrder(members, properties)), 2)

	// Nothing is recorded for a single member, for a table with a member that getNamedMembers
	// left out, or for a member that is not a plain value member.
	assert.Assert(t, createMemberOrder(ast.SymbolTable{"zDeclared": members["zDeclared"]}, properties[:1]) == nil)
	withTypeOnly := f.members()
	withTypeOnly["typeOnly"] = &ast.Symbol{Name: "typeOnly", Flags: ast.SymbolFlagsInterface}
	assert.Assert(t, createMemberOrder(withTypeOnly, c.getNamedMembers(withTypeOnly, f.container)) == nil)
	typeOnly := f.members()
	typeOnly["zDeclared"].Flags = ast.SymbolFlagsInterface
	assert.Assert(t, createMemberOrder(typeOnly, []*ast.Symbol{typeOnly["zDeclared"], typeOnly["aInherited"]}) == nil)
}
