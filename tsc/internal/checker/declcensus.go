package checker

import (
	"cmp"
	"fmt"
	"slices"
	"time"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/scanner"
)

// Declaration census
//
// Research instrumentation for the G1 gate (research/agents/20260913T150256-fable-5.1.md,
// section 1). With TSGO_DECL_CENSUS set, every miss of a memo slot that holds a declaration-level
// result (declared types, resolved symbol types, members, base types, signatures, return types,
// object/conditional/alias instantiations, variances) records one frame: a 128-bit provenance
// fingerprint that identifies the same computation in every checker, the exclusive wall time and
// counter deltas of the computation, and flags saying why its result could not be shared. Hits
// on the same slots record dependency edges. The compiler package joins the frames of all
// checkers once they have finished (compiler/declcensus.go).
//
// A nil *DeclCensus records nothing: every hook is a method with a nil-receiver check, so the
// checker pays one nil compare per hooked miss, hit, diagnostic and instantiation when the
// census is off.

// DeclCensusKind names the memo slot a frame computed.
type DeclCensusKind uint8

const (
	DeclCensusKindDeclaredType             DeclCensusKind = iota // getDeclaredTypeOfClassOrInterface, getDeclaredTypeOfTypeAlias, getDeclaredTypeOfEnum
	DeclCensusKindResolvedType                                   // getTypeOfVariableOrParameterOrProperty, getTypeOfFuncClassEnumModule
	DeclCensusKindInstantiatedType                               // getTypeOfInstantiatedSymbol
	DeclCensusKindMembers                                        // resolveStructuredTypeMembers of a declared, anonymous, mapped, union or intersection type
	DeclCensusKindReferenceMembers                               // resolveStructuredTypeMembers of a type reference
	DeclCensusKindBaseTypes                                      // getBaseTypes
	DeclCensusKindDeclaredMembers                                // resolveDeclaredMembers
	DeclCensusKindResolvedSignature                              // getSignatureFromDeclaration
	DeclCensusKindResolvedReturnType                             // getReturnTypeOfSignature
	DeclCensusKindObjectInstantiation                            // getObjectTypeInstantiation
	DeclCensusKindConditionalInstantiation                       // getConditionalTypeInstantiation
	DeclCensusKindAliasInstantiation                             // getTypeAliasInstantiation
	DeclCensusKindVariances                                      // getVariancesWorker
	DeclCensusKindCount
)

var declCensusKindNames = [DeclCensusKindCount]string{
	"declaredType", "resolvedType", "instantiatedType", "members", "referenceMembers", "baseTypes",
	"declaredMembers", "resolvedSignature", "resolvedReturnType", "objectInstantiation",
	"conditionalInstantiation", "aliasInstantiation", "variances",
}

func (k DeclCensusKind) String() string {
	return declCensusKindNames[k]
}

// DeclCensusFlags describe a frame. The poison flags say why the result of the frame could not
// have been shared between checkers, and the source-key flag says that the computation depends
// on a declaration outside the declaration files. Both are inherited: at pop a frame passes
// them to the enclosing frame, and a frame that reads a memo slot on a hit takes them from the
// frame that filled the slot, so a frame carries the flags of everything its result was
// computed from. Generic is a property of the frame's own declaration.
type DeclCensusFlags uint8

const (
	DeclCensusFlagsSrcExpr   DeclCensusFlags = 1 << iota // an expression of a non-declaration file was checked inside the frame
	DeclCensusFlagsDiag                                  // a diagnostic was emitted inside the frame
	DeclCensusFlagsGuard                                 // an instantiation depth/count, conditional tail or relation budget guard fired inside the frame
	DeclCensusFlagsCycle                                 // a type resolution cycle or partially resolved members were observed inside the frame
	DeclCensusFlagsUnkeyable                             // part of the key material has no cross-checker identity
	DeclCensusFlagsSourceKey                             // part of the key material is a node or symbol of a non-declaration file
	DeclCensusFlagsGeneric                               // the computed declaration or signature has type parameters

	DeclCensusPoisonFlags    = DeclCensusFlagsSrcExpr | DeclCensusFlagsDiag | DeclCensusFlagsGuard | DeclCensusFlagsCycle | DeclCensusFlagsUnkeyable
	declCensusInheritedFlags = DeclCensusPoisonFlags | DeclCensusFlagsSourceKey
)

var declCensusFlagNames = [...]struct {
	flag DeclCensusFlags
	name string
}{
	{DeclCensusFlagsSrcExpr, "srcExpr"},
	{DeclCensusFlagsDiag, "diag"},
	{DeclCensusFlagsGuard, "guard"},
	{DeclCensusFlagsCycle, "cycle"},
	{DeclCensusFlagsUnkeyable, "unkeyable"},
	{DeclCensusFlagsSourceKey, "sourceKey"},
	{DeclCensusFlagsGeneric, "generic"},
}

// Names returns the names of the set flags.
func (f DeclCensusFlags) Names() []string {
	var names []string
	for _, fn := range declCensusFlagNames {
		if f&fn.flag != 0 {
			names = append(names, fn.name)
		}
	}
	return names
}

// DeclCensusFrame is one memo-slot miss. Start and End are nanoseconds since the census epoch.
// Types, Symbols, Signatures and Instantiations are the objects created by the frame itself
// (frames nested inside it subtracted); InclInstantiations and Depth are the instantiations
// and the maximum instantiation depth reached, relative to entry, of the whole computation,
// which is what the instantiation guards see.
type DeclCensusFrame struct {
	Key                CacheHashKey
	Parent             int32 // index of the enclosing open frame, -1 for a root frame
	Kind               DeclCensusKind
	Flags              DeclCensusFlags
	Start              int64
	End                int64
	ExclNs             int64
	Types              uint32
	Symbols            uint32
	Signatures         uint32
	Instantiations     uint32
	InclInstantiations uint32
	Depth              uint32
	Anchor             *ast.Node // declaration behind the key, for the description
	Name               string    // symbol name behind the key, for the description
}

// Description names the computation for a reader: the symbol name and the file:line of its
// declaration, whichever are known.
func (f *DeclCensusFrame) Description() string {
	if f.Anchor == nil {
		return f.Name
	}
	file := ast.GetSourceFileOfNode(f.Anchor)
	if file == nil {
		return f.Name
	}
	location := fmt.Sprintf("%s:%d", file.FileName(), scanner.ComputeLineOfPosition(file.ECMALineMap(), f.Anchor.Pos())+1)
	if f.Name == "" {
		return location
	}
	return f.Name + " " + location
}

// declCensusMaxFrames bounds the memory of one checker's census (about 100 bytes per frame);
// past it the census stops recording and reports itself as truncated.
const declCensusMaxFrames = 4_000_000

// declCensusOpenFrame is the bookkeeping of a frame while it is on the stack.
type declCensusOpenFrame struct {
	frame         int32
	entryDepth    uint32    // instantiationDepth when the frame was pushed
	savedMaxDepth uint32    // running maximum depth of the enclosing frame before the push
	childNs       int64     // inclusive ns of the direct child frames popped so far
	childCounters [4]uint32 // inclusive counter deltas of the direct child frames popped so far
}

// DeclCensus records the frames of one checker. Only the goroutine that holds the checker
// touches it, so it needs no locking; the join over checkers happens after they are done.
type DeclCensus struct {
	c              *Checker
	epoch          time.Time
	initTypes      TypeId      // types created by NewChecker: the same ids in every checker
	initSignatures SignatureId // signatures created by NewChecker: likewise
	Frames         []DeclCensusFrame
	Truncated      bool
	stack          []declCensusOpenFrame
	maxDepth       uint32                 // maximum instantiationDepth since the innermost frame was pushed
	frameByKey     map[CacheHashKey]int32 // first frame of every key, the target of hit edges
	hitEdges       map[uint64]struct{}    // parent frame index << 32 | child frame index

	typeFingerprints      []fingerprintMemo // indexed by type id
	signatureFingerprints []fingerprintMemo // indexed by signature id
	symbolFingerprints    core.PagedLinkStore[fingerprintMemo]
}

// EnableDeclCensus attaches a census to the checker. It is called right after NewChecker, so
// that the types and signatures the checker has created so far are exactly those every
// checker creates in the same order.
func (c *Checker) EnableDeclCensus(epoch time.Time) *DeclCensus {
	c.census = &DeclCensus{
		c:              c,
		epoch:          epoch,
		initTypes:      TypeId(c.TypeCount),
		initSignatures: SignatureId(c.SignatureCount),
		frameByKey:     make(map[CacheHashKey]int32),
		hitEdges:       make(map[uint64]struct{}),
	}
	return c.census
}

// HitEdges returns the dependencies recorded on hit paths as (parent frame, child frame) index
// pairs, one per pair.
func (cs *DeclCensus) HitEdges() [][2]int32 {
	edges := make([][2]int32, 0, len(cs.hitEdges))
	for edge := range cs.hitEdges {
		edges = append(edges, [2]int32{int32(edge >> 32), int32(uint32(edge))})
	}
	slices.SortFunc(edges, func(a, b [2]int32) int {
		return cmp.Or(cmp.Compare(a[0], b[0]), cmp.Compare(a[1], b[1]))
	})
	return edges
}

func (cs *DeclCensus) now() int64 {
	return time.Since(cs.epoch).Nanoseconds()
}

// Miss hooks. Each push opens a frame for a memo-slot miss and returns its index (-1 when
// nothing is recorded); the matching pop closes it once the slot is written. Every hook makes
// one call past its nil check, to a worker too large to be inlined into it, so that the compiler
// inlines the hook itself and the checker pays one compare when the census is off.

func (cs *DeclCensus) pushSymbol(kind DeclCensusKind, symbol *ast.Symbol) int {
	if cs == nil {
		return -1
	}
	return cs.openSymbolFrame(kind, symbol)
}

func (cs *DeclCensus) pushMembers(t *Type) int {
	if cs == nil {
		return -1
	}
	return cs.openMembersFrame(t)
}

func (cs *DeclCensus) pushType(kind DeclCensusKind, t *Type) int {
	if cs == nil {
		return -1
	}
	return cs.openTypeFrame(kind, t)
}

func (cs *DeclCensus) pushSignatureDeclaration(declaration *ast.Node) int {
	if cs == nil {
		return -1
	}
	return cs.openNodeFrame(DeclCensusKindResolvedSignature, declaration)
}

func (cs *DeclCensus) pushReturnType(sig *Signature) int {
	if cs == nil {
		return -1
	}
	return cs.openSignatureFrame(DeclCensusKindResolvedReturnType, sig)
}

func (cs *DeclCensus) pushObjectInstantiation(target *Type, typeArguments []*Type, alias *TypeAlias, singleSignature bool) int {
	if cs == nil {
		return -1
	}
	return cs.openObjectInstantiationFrame(target, typeArguments, alias, singleSignature)
}

// pushConditionalInstantiation takes a single type argument apart from the argument slice
// because getConditionalTypeInstantiation keeps it out of a slice on its hit path; exactly one
// of the two is set.
func (cs *DeclCensus) pushConditionalInstantiation(root *ConditionalRoot, singleArgument *Type, typeArguments []*Type, alias *TypeAlias, forConstraint bool) int {
	if cs == nil {
		return -1
	}
	return cs.openConditionalInstantiationFrame(root, singleArgument, typeArguments, alias, forConstraint)
}

func (cs *DeclCensus) pushAliasInstantiation(symbol *ast.Symbol, typeArguments []*Type, alias *TypeAlias) int {
	if cs == nil {
		return -1
	}
	return cs.openAliasInstantiationFrame(symbol, typeArguments, alias)
}

// pop closes the frame opened by the matching push. generic says whether the computed
// declaration or signature has type parameters, which the first sound subset excludes.
func (cs *DeclCensus) pop(frame int, generic bool) {
	if cs == nil || frame < 0 {
		return
	}
	cs.closeFrame(frame, generic)
}

// Hit hooks. A hit on a memo slot inside an open frame is a dependency of that frame on the
// computation that filled the slot.

func (cs *DeclCensus) hitSymbol(kind DeclCensusKind, symbol *ast.Symbol) {
	if cs == nil {
		return
	}
	cs.recordSymbolHit(kind, symbol)
}

func (cs *DeclCensus) hitMembers(t *Type) {
	if cs == nil {
		return
	}
	cs.recordMembersHit(t)
}

func (cs *DeclCensus) hitType(kind DeclCensusKind, t *Type) {
	if cs == nil {
		return
	}
	cs.recordTypeHit(kind, t)
}

func (cs *DeclCensus) hitSignatureDeclaration(declaration *ast.Node) {
	if cs == nil {
		return
	}
	cs.recordNodeHit(DeclCensusKindResolvedSignature, declaration)
}

func (cs *DeclCensus) hitReturnType(sig *Signature) {
	if cs == nil {
		return
	}
	cs.recordSignatureHit(DeclCensusKindResolvedReturnType, sig)
}

func (cs *DeclCensus) hitObjectInstantiation(target *Type, typeArguments []*Type, alias *TypeAlias, singleSignature bool) {
	if cs == nil {
		return
	}
	cs.recordObjectInstantiationHit(target, typeArguments, alias, singleSignature)
}

func (cs *DeclCensus) hitConditionalInstantiation(root *ConditionalRoot, singleArgument *Type, typeArguments []*Type, alias *TypeAlias, forConstraint bool) {
	if cs == nil {
		return
	}
	cs.recordConditionalInstantiationHit(root, singleArgument, typeArguments, alias, forConstraint)
}

func (cs *DeclCensus) hitAliasInstantiation(symbol *ast.Symbol, typeArguments []*Type, alias *TypeAlias) {
	if cs == nil {
		return
	}
	cs.recordAliasInstantiationHit(symbol, typeArguments, alias)
}

// Flag hooks.

// flag poisons the innermost open frame; pop carries the flag to the enclosing frames.
func (cs *DeclCensus) flag(f DeclCensusFlags) {
	if cs == nil {
		return
	}
	cs.markFrame(f)
}

// sourceExpression records that an expression is being checked inside the innermost open frame
// when the expression belongs to a file that is not a declaration file.
func (cs *DeclCensus) sourceExpression(node *ast.Node) {
	if cs == nil {
		return
	}
	cs.markSourceExpression(node)
}

// instantiation records the depth of a type instantiation for the Depth of the open frames.
func (cs *DeclCensus) instantiation(depth uint32) {
	if cs != nil && depth > cs.maxDepth {
		cs.maxDepth = depth
	}
}

// Recording.

func (cs *DeclCensus) openSymbolFrame(kind DeclCensusKind, symbol *ast.Symbol) int {
	return cs.openFrame(censusFrameKey(kind, cs.c.censusSymbolFingerprint(symbol)), censusSymbolAnchor(symbol), symbol.Name)
}

func (cs *DeclCensus) openTypeFrame(kind DeclCensusKind, t *Type) int {
	anchor, name := censusTypeAnchor(t)
	return cs.openFrame(censusFrameKey(kind, cs.c.censusTypeFingerprint(t)), anchor, name)
}

func (cs *DeclCensus) openMembersFrame(t *Type) int {
	return cs.openTypeFrame(censusMembersKind(t), t)
}

func (cs *DeclCensus) openNodeFrame(kind DeclCensusKind, node *ast.Node) int {
	return cs.openFrame(censusFrameKey(kind, censusNodeFingerprint(node)), node, censusNodeName(node))
}

func (cs *DeclCensus) openSignatureFrame(kind DeclCensusKind, sig *Signature) int {
	var name string
	if sig.declaration != nil {
		name = censusNodeName(sig.declaration)
	}
	return cs.openFrame(censusFrameKey(kind, cs.c.censusSignatureFingerprint(sig)), sig.declaration, name)
}

func (cs *DeclCensus) openObjectInstantiationFrame(target *Type, typeArguments []*Type, alias *TypeAlias, singleSignature bool) int {
	anchor, name := censusTypeAnchor(target)
	return cs.openFrame(cs.objectInstantiationKey(target, typeArguments, alias, singleSignature), anchor, name)
}

func (cs *DeclCensus) openConditionalInstantiationFrame(root *ConditionalRoot, singleArgument *Type, typeArguments []*Type, alias *TypeAlias, forConstraint bool) int {
	anchor, name := censusConditionalAnchor(root)
	return cs.openFrame(cs.conditionalInstantiationKey(root, singleArgument, typeArguments, alias, forConstraint), anchor, name)
}

func (cs *DeclCensus) openAliasInstantiationFrame(symbol *ast.Symbol, typeArguments []*Type, alias *TypeAlias) int {
	return cs.openFrame(cs.aliasInstantiationKey(symbol, typeArguments, alias), censusSymbolAnchor(symbol), symbol.Name)
}

// openFrame opens a frame. The counters are stored at entry and turned into deltas by
// closeFrame.
func (cs *DeclCensus) openFrame(key censusKey, anchor *ast.Node, name string) int {
	if len(cs.Frames) >= declCensusMaxFrames {
		cs.Truncated = true
		return -1
	}
	c := cs.c
	index := int32(len(cs.Frames))
	hash := key.hash
	if key.flags&DeclCensusFlagsUnkeyable != 0 {
		// Without a cross-checker identity two unkeyable computations must never be taken for
		// the same one: the key is made unique to this frame.
		hash = censusUnkeyableFrameKey(hash, c.id, index)
	}
	parent := int32(-1)
	if n := len(cs.stack); n > 0 {
		parent = cs.stack[n-1].frame
	}
	cs.Frames = append(cs.Frames, DeclCensusFrame{
		Key:            hash,
		Parent:         parent,
		Kind:           key.kind,
		Flags:          key.flags,
		Types:          c.TypeCount,
		Symbols:        c.SymbolCount,
		Signatures:     c.SignatureCount,
		Instantiations: c.TotalInstantiationCount,
		Anchor:         anchor,
		Name:           name,
		Start:          cs.now(),
	})
	cs.stack = append(cs.stack, declCensusOpenFrame{frame: index, entryDepth: c.instantiationDepth, savedMaxDepth: cs.maxDepth})
	cs.maxDepth = c.instantiationDepth
	if _, seen := cs.frameByKey[hash]; !seen {
		cs.frameByKey[hash] = index
	}
	return int(index)
}

// closeFrame turns the entry counters into exclusive deltas, charges the frame's inclusive cost
// to its parent and passes its poison flags up.
func (cs *DeclCensus) closeFrame(frame int, generic bool) {
	n := len(cs.stack) - 1
	open := cs.stack[n]
	if open.frame != int32(frame) {
		panic("declaration census: frames are not nested")
	}
	cs.stack = cs.stack[:n]
	c := cs.c
	f := &cs.Frames[frame]
	f.End = cs.now()
	inclusiveNs := f.End - f.Start
	f.ExclNs = inclusiveNs - open.childNs
	inclusive := [4]uint32{c.TypeCount - f.Types, c.SymbolCount - f.Symbols, c.SignatureCount - f.Signatures, c.TotalInstantiationCount - f.Instantiations}
	f.Types = inclusive[0] - open.childCounters[0]
	f.Symbols = inclusive[1] - open.childCounters[1]
	f.Signatures = inclusive[2] - open.childCounters[2]
	f.Instantiations = inclusive[3] - open.childCounters[3]
	f.InclInstantiations = inclusive[3]
	f.Depth = cs.maxDepth - open.entryDepth
	if generic {
		f.Flags |= DeclCensusFlagsGeneric
	}
	if n > 0 {
		parent := &cs.stack[n-1]
		parent.childNs += inclusiveNs
		for i, delta := range inclusive {
			parent.childCounters[i] += delta
		}
		cs.Frames[parent.frame].Flags |= f.Flags & declCensusInheritedFlags
	}
	cs.maxDepth = max(open.savedMaxDepth, cs.maxDepth)
}

func (cs *DeclCensus) recordSymbolHit(kind DeclCensusKind, symbol *ast.Symbol) {
	cs.recordHit(kind, cs.c.censusSymbolFingerprint(symbol))
}

func (cs *DeclCensus) recordMembersHit(t *Type) {
	if t.objectFlags&ObjectFlagsUnresolvedMembers != 0 {
		// The members are read while the base types of the type are still being resolved:
		// the reader gets a partial member set.
		cs.markFrame(DeclCensusFlagsCycle)
	}
	cs.recordTypeHit(censusMembersKind(t), t)
}

func (cs *DeclCensus) recordTypeHit(kind DeclCensusKind, t *Type) {
	cs.recordHit(kind, cs.c.censusTypeFingerprint(t))
}

func (cs *DeclCensus) recordNodeHit(kind DeclCensusKind, node *ast.Node) {
	cs.recordHit(kind, censusNodeFingerprint(node))
}

func (cs *DeclCensus) recordSignatureHit(kind DeclCensusKind, sig *Signature) {
	cs.recordHit(kind, cs.c.censusSignatureFingerprint(sig))
}

func (cs *DeclCensus) recordObjectInstantiationHit(target *Type, typeArguments []*Type, alias *TypeAlias, singleSignature bool) {
	cs.recordKeyHit(cs.objectInstantiationKey(target, typeArguments, alias, singleSignature))
}

func (cs *DeclCensus) recordConditionalInstantiationHit(root *ConditionalRoot, singleArgument *Type, typeArguments []*Type, alias *TypeAlias, forConstraint bool) {
	cs.recordKeyHit(cs.conditionalInstantiationKey(root, singleArgument, typeArguments, alias, forConstraint))
}

func (cs *DeclCensus) recordAliasInstantiationHit(symbol *ast.Symbol, typeArguments []*Type, alias *TypeAlias) {
	cs.recordKeyHit(cs.aliasInstantiationKey(symbol, typeArguments, alias))
}

func (cs *DeclCensus) recordHit(kind DeclCensusKind, fp fingerprint) {
	cs.recordKeyHit(censusFrameKey(kind, fp))
}

// recordKeyHit makes the innermost open frame depend on the frame that computed the key: it
// inherits that frame's poison and source-key flags and gets a hit edge to it. Nothing is
// recorded when this checker has no frame for the key (the slot was filled before the census
// was attached or by an unhooked producer) or when the frame is still open (a hit on an open
// frame is structural recursion, not a dependency on earlier work).
func (cs *DeclCensus) recordKeyHit(key censusKey) {
	n := len(cs.stack)
	if n == 0 {
		return
	}
	parent := cs.stack[n-1].frame
	if key.flags&DeclCensusFlagsUnkeyable != 0 {
		// A dependency without a cross-checker identity leaves the reader without one too.
		cs.Frames[parent].Flags |= DeclCensusFlagsUnkeyable
		return
	}
	child, ok := cs.frameByKey[key.hash]
	if !ok || cs.Frames[child].End == 0 {
		return
	}
	cs.Frames[parent].Flags |= cs.Frames[child].Flags & declCensusInheritedFlags
	cs.hitEdges[uint64(parent)<<32|uint64(child)] = struct{}{}
}

func (cs *DeclCensus) markFrame(f DeclCensusFlags) {
	if n := len(cs.stack); n > 0 {
		cs.Frames[cs.stack[n-1].frame].Flags |= f
	}
}

// markSourceExpression is kept out of line so that sourceExpression, on the path of every
// checked expression, stays small enough to be inlined.
//
//go:noinline
func (cs *DeclCensus) markSourceExpression(node *ast.Node) {
	n := len(cs.stack)
	if n == 0 {
		return
	}
	f := &cs.Frames[cs.stack[n-1].frame]
	if f.Flags&DeclCensusFlagsSrcExpr != 0 {
		return
	}
	if file := ast.GetSourceFileOfNode(node); file == nil || !file.IsDeclarationFile {
		f.Flags |= DeclCensusFlagsSrcExpr
	}
}

// Keys and descriptions.

// censusKey is a frame key: the hash of the kind and the key material, with the flags of the
// material.
type censusKey struct {
	kind  DeclCensusKind
	hash  CacheHashKey
	flags DeclCensusFlags
}

func censusFrameKey(kind DeclCensusKind, fp fingerprint) censusKey {
	var b keyBuilder
	b.writeByte(byte(kind))
	b.writeUint64(fp.key.Hi)
	b.writeUint64(fp.key.Lo)
	return censusKey{kind: kind, hash: b.hash(), flags: fp.flags}
}

func censusUnkeyableFrameKey(key CacheHashKey, checker uint32, frame int32) CacheHashKey {
	var b keyBuilder
	b.writeUint64(key.Hi)
	b.writeUint64(key.Lo)
	b.writeUint32(checker)
	b.writeUint32(uint32(frame))
	return b.hash()
}

func (cs *DeclCensus) objectInstantiationKey(target *Type, typeArguments []*Type, alias *TypeAlias, singleSignature bool) censusKey {
	c := cs.c
	var b fingerprintBuilder
	b.writeByte(byte(DeclCensusKindObjectInstantiation))
	b.writeFingerprint(c.censusTypeFingerprint(target))
	b.writeTypeFingerprints(c, typeArguments)
	b.writeAliasFingerprint(c, alias)
	if singleSignature {
		b.writeByte('!')
	}
	return b.key(DeclCensusKindObjectInstantiation)
}

func (cs *DeclCensus) conditionalInstantiationKey(root *ConditionalRoot, singleArgument *Type, typeArguments []*Type, alias *TypeAlias, forConstraint bool) censusKey {
	c := cs.c
	var b fingerprintBuilder
	b.writeByte(byte(DeclCensusKindConditionalInstantiation))
	b.writeNodeIdentity(root.node.AsNode())
	if singleArgument != nil {
		b.writeInt(1)
		b.writeFingerprint(c.censusTypeFingerprint(singleArgument))
	} else {
		b.writeTypeFingerprints(c, typeArguments)
	}
	b.writeAliasFingerprint(c, alias)
	if forConstraint {
		b.writeByte('!')
	}
	return b.key(DeclCensusKindConditionalInstantiation)
}

func (cs *DeclCensus) aliasInstantiationKey(symbol *ast.Symbol, typeArguments []*Type, alias *TypeAlias) censusKey {
	c := cs.c
	var b fingerprintBuilder
	b.writeByte(byte(DeclCensusKindAliasInstantiation))
	b.writeFingerprint(c.censusSymbolFingerprint(symbol))
	b.writeTypeFingerprints(c, typeArguments)
	b.writeAliasFingerprint(c, alias)
	return b.key(DeclCensusKindAliasInstantiation)
}

// censusMembersKind separates the members of type references, which are instantiation work,
// from the members of declared and structural types.
func censusMembersKind(t *Type) DeclCensusKind {
	if t.flags&TypeFlagsObject != 0 && t.objectFlags&ObjectFlagsReference != 0 {
		return DeclCensusKindReferenceMembers
	}
	return DeclCensusKindMembers
}

func censusSymbolAnchor(symbol *ast.Symbol) *ast.Node {
	if symbol.ValueDeclaration != nil {
		return symbol.ValueDeclaration
	}
	if len(symbol.Declarations) != 0 {
		return symbol.Declarations[0]
	}
	return nil
}

func censusTypeAnchor(t *Type) (*ast.Node, string) {
	switch {
	case t.symbol != nil:
		return censusSymbolAnchor(t.symbol), t.symbol.Name
	case t.alias != nil:
		return censusSymbolAnchor(t.alias.symbol), t.alias.symbol.Name
	}
	return nil, ""
}

func censusConditionalAnchor(root *ConditionalRoot) (*ast.Node, string) {
	if root.alias != nil {
		return root.node.AsNode(), root.alias.symbol.Name
	}
	return root.node.AsNode(), ""
}

func censusNodeName(node *ast.Node) string {
	if symbol := node.Symbol(); symbol != nil {
		return symbol.Name
	}
	return ""
}

// Fingerprints
//
// A fingerprint is the 128-bit provenance key of a symbol, type, mapper, signature or node: the
// same value in every checker for the same computation, built from global ids (binder symbols,
// nodes) and structurally over the checker-local objects (types, mappers, signatures, transient
// symbols), the way CacheHashKey is built from local ids. It carries what the census needs to
// know about its material: whether some part has no cross-checker identity, and whether some
// part belongs to a file that is not a declaration file. Fingerprints are memoised per checker
// by type id, signature id and global symbol id. The fingerprint is a provenance approximation
// for the census, not a cache key: two differently constructed objects with the same
// fingerprint are counted as the same computation.

type fingerprint struct {
	key   CacheHashKey
	flags DeclCensusFlags // DeclCensusFlagsUnkeyable, DeclCensusFlagsSourceKey
}

var unkeyableFingerprint = fingerprint{flags: DeclCensusFlagsUnkeyable}

// fingerprintMemo is a memoised fingerprint; the zero state means not computed yet.
type fingerprintMemo struct {
	fingerprint
	state uint8
}

const (
	fingerprintInProgress uint8 = iota + 1
	fingerprintDone
)

// cached returns the memoised fingerprint, or reports that it has to be computed. A
// fingerprint that is still being computed recursively needs itself and resolves to unkeyable
// instead of recursing forever.
func (m *fingerprintMemo) cached() (fingerprint, bool) {
	switch m.state {
	case fingerprintInProgress:
		return unkeyableFingerprint, true
	case fingerprintDone:
		return m.fingerprint, true
	default:
		return fingerprint{}, false
	}
}

// fingerprintBuilder hashes the parts of a fingerprint like keyBuilder hashes cache keys and
// accumulates the flags of the parts.
type fingerprintBuilder struct {
	keyBuilder
	flags DeclCensusFlags
}

func (b *fingerprintBuilder) fingerprint() fingerprint {
	return fingerprint{key: b.hash(), flags: b.flags}
}

func (b *fingerprintBuilder) key(kind DeclCensusKind) censusKey {
	return censusKey{kind: kind, hash: b.hash(), flags: b.flags}
}

func (b *fingerprintBuilder) writeFingerprint(fp fingerprint) {
	b.writeUint64(fp.key.Hi)
	b.writeUint64(fp.key.Lo)
	b.flags |= fp.flags
}

func (b *fingerprintBuilder) writeTypeFingerprints(c *Checker, types []*Type) {
	b.writeInt(len(types))
	for _, t := range types {
		b.writeFingerprint(c.censusTypeFingerprint(t))
	}
}

func (b *fingerprintBuilder) writeAliasFingerprint(c *Checker, alias *TypeAlias) {
	if alias == nil {
		b.writeByte(0)
		return
	}
	b.writeByte(1)
	b.writeFingerprint(c.censusSymbolFingerprint(alias.symbol))
	b.writeTypeFingerprints(c, alias.typeArguments)
}

// writeNodeIdentity writes the global node id; a node outside a declaration file makes the
// fingerprint source-anchored.
func (b *fingerprintBuilder) writeNodeIdentity(node *ast.Node) {
	b.writeUint64(uint64(ast.GetNodeId(node)))
	if !censusIsDeclarationNode(node) {
		b.flags |= DeclCensusFlagsSourceKey
	}
}

// markDeclarations makes the fingerprint source-anchored when any of the declarations is
// outside a declaration file.
func (b *fingerprintBuilder) markDeclarations(declarations []*ast.Node) {
	for _, d := range declarations {
		if !censusIsDeclarationNode(d) {
			b.flags |= DeclCensusFlagsSourceKey
			return
		}
	}
}

func censusIsDeclarationNode(node *ast.Node) bool {
	file := ast.GetSourceFileOfNode(node)
	return file != nil && file.IsDeclarationFile
}

func censusNodeFingerprint(node *ast.Node) fingerprint {
	var b fingerprintBuilder
	b.writeByte('n')
	b.writeNodeIdentity(node)
	return b.fingerprint()
}

// censusSymbolFingerprint is the provenance fingerprint of a symbol. Binder symbols are shared
// by every checker, so their global id is their identity. Transient symbols are checker-local
// and are keyed by what they were made from: instantiated symbols by target and mapper,
// synthetic and mapped properties by containing type and name, merged, cloned and late-bound
// symbols by every declaration (so an augmentation in another file changes the key), and
// reverse-mapped properties by nothing, because they are inferred from expressions.
func (c *Checker) censusSymbolFingerprint(symbol *ast.Symbol) fingerprint {
	memo := c.census.symbolFingerprints.Get(uint64(ast.GetSymbolId(symbol)))
	if fp, ok := memo.cached(); ok {
		return fp
	}
	memo.state = fingerprintInProgress
	fp := c.censusSymbolFingerprintWorker(symbol)
	// The store hands out pointers into fixed pages, so memo is still valid after the recursion.
	memo.fingerprint = fp
	memo.state = fingerprintDone
	return fp
}

func (c *Checker) censusSymbolFingerprintWorker(symbol *ast.Symbol) fingerprint {
	var b fingerprintBuilder
	links := c.valueSymbolLinks.TryGet(symbol)
	switch {
	case symbol.Flags&ast.SymbolFlagsTransient == 0:
		b.writeByte('S')
		b.writeUint64(uint64(ast.GetSymbolId(symbol)))
		b.markDeclarations(symbol.Declarations)
	case symbol.CheckFlags&ast.CheckFlagsReverseMapped != 0:
		return unkeyableFingerprint
	case symbol.CheckFlags&ast.CheckFlagsInstantiated != 0 && links != nil:
		b.writeByte('I')
		b.writeFingerprint(c.censusSymbolFingerprint(links.target))
		b.writeFingerprint(c.censusMapperFingerprint(links.mapper))
	case links != nil && links.containingType != nil:
		b.writeByte('P')
		b.writeFingerprint(c.censusTypeFingerprint(links.containingType))
		b.writeString(symbol.Name)
		if mapped := c.mappedSymbolLinks.TryGet(symbol); mapped != nil && mapped.keyType != nil {
			b.writeFingerprint(c.censusTypeFingerprint(mapped.keyType))
		}
	case len(symbol.Declarations) != 0:
		b.writeByte('M')
		b.writeUint32(uint32(symbol.Flags))
		ids := make([]uint64, len(symbol.Declarations))
		for i, d := range symbol.Declarations {
			ids[i] = uint64(ast.GetNodeId(d))
		}
		slices.Sort(ids)
		for _, id := range ids {
			b.writeUint64(id)
		}
		b.markDeclarations(symbol.Declarations)
	default:
		// Symbols the checker makes without declarations (unknownSymbol, globalThis, index
		// symbols, ...): the same name and flags in every checker.
		b.writeByte('N')
		b.writeString(symbol.Name)
		b.writeUint32(uint32(symbol.Flags))
	}
	return b.fingerprint()
}

// censusTypeFingerprint is the provenance fingerprint of a type (section 1.4 of the report):
// structural over the type's parts, with the type's declaration and symbol as the anchors.
func (c *Checker) censusTypeFingerprint(t *Type) fingerprint {
	cs := c.census
	if int(t.id) >= len(cs.typeFingerprints) {
		// Grown to the current count at once; fingerprinting creates no types.
		cs.typeFingerprints = slices.Grow(cs.typeFingerprints, int(c.TypeCount)+1-len(cs.typeFingerprints))[:int(c.TypeCount)+1]
	}
	if fp, ok := cs.typeFingerprints[t.id].cached(); ok {
		return fp
	}
	cs.typeFingerprints[t.id].state = fingerprintInProgress
	fp := c.censusTypeFingerprintWorker(t)
	// Indexed again rather than through a pointer taken before the recursion, which is what a
	// grown slice would invalidate.
	cs.typeFingerprints[t.id] = fingerprintMemo{fingerprint: fp, state: fingerprintDone}
	return fp
}

func (c *Checker) censusTypeFingerprintWorker(t *Type) fingerprint {
	var b fingerprintBuilder
	switch {
	case t.id <= c.census.initTypes:
		// Created by NewChecker in a fixed order: the id is the same in every checker.
		b.writeByte('0')
		b.writeUint32(uint32(t.id))
	case t.flags&TypeFlagsIntrinsic != 0:
		b.writeByte('i')
		b.writeString(t.AsIntrinsicType().intrinsicName)
		b.writeUint32(uint32(t.flags))
		b.writeUint32(uint32(t.objectFlags))
	case t.flags&(TypeFlagsLiteral|TypeFlagsEnum) != 0:
		d := t.AsLiteralType()
		b.writeByte('l')
		b.writeUint32(uint32(t.flags))
		if d.value != nil {
			b.writeString(ValueToString(d.value))
		}
		if d.freshType == t {
			b.writeByte('f')
		}
		if t.symbol != nil {
			b.writeFingerprint(c.censusSymbolFingerprint(t.symbol))
		}
	case t.flags&TypeFlagsUniqueESSymbol != 0:
		b.writeByte('u')
		b.writeFingerprint(c.censusSymbolFingerprint(t.symbol))
	case t.flags&TypeFlagsTypeParameter != 0:
		return c.censusTypeParameterFingerprint(t)
	case t.flags&TypeFlagsObject != 0:
		return c.censusObjectTypeFingerprint(t)
	case t.flags&TypeFlagsUnion != 0:
		// Constituents stay in the checker's order and the origin is part of the identity:
		// CompareTypes observes both, so an order-insensitive key would merge distinct unions.
		u := t.AsUnionType()
		b.writeByte('|')
		b.writeTypeFingerprints(c, u.types)
		if u.origin != nil {
			b.writeByte('o')
			b.writeFingerprint(c.censusTypeFingerprint(u.origin))
		}
		b.writeAliasFingerprint(c, t.alias)
	case t.flags&TypeFlagsIntersection != 0:
		b.writeByte('&')
		b.writeTypeFingerprints(c, t.AsIntersectionType().types)
		b.writeAliasFingerprint(c, t.alias)
	case t.flags&TypeFlagsIndex != 0:
		d := t.AsIndexType()
		b.writeByte('k')
		b.writeFingerprint(c.censusTypeFingerprint(d.target))
		b.writeUint32(uint32(d.indexFlags))
	case t.flags&TypeFlagsIndexedAccess != 0:
		d := t.AsIndexedAccessType()
		b.writeByte('x')
		b.writeFingerprint(c.censusTypeFingerprint(d.objectType))
		b.writeFingerprint(c.censusTypeFingerprint(d.indexType))
		b.writeUint32(uint32(d.accessFlags))
	case t.flags&TypeFlagsConditional != 0:
		d := t.AsConditionalType()
		b.writeByte('c')
		b.writeNodeIdentity(d.root.node.AsNode())
		b.writeFingerprint(c.censusMapperFingerprint(d.mapper))
		b.writeAliasFingerprint(c, t.alias)
	case t.flags&TypeFlagsSubstitution != 0:
		d := t.AsSubstitutionType()
		b.writeByte('s')
		b.writeFingerprint(c.censusTypeFingerprint(d.baseType))
		b.writeFingerprint(c.censusTypeFingerprint(d.constraint))
	case t.flags&TypeFlagsTemplateLiteral != 0:
		d := t.AsTemplateLiteralType()
		b.writeByte('t')
		b.writeInt(len(d.texts))
		for _, text := range d.texts {
			b.writeInt(len(text))
			b.writeString(text)
		}
		b.writeTypeFingerprints(c, d.types)
	case t.flags&TypeFlagsStringMapping != 0:
		b.writeByte('m')
		b.writeFingerprint(c.censusSymbolFingerprint(t.symbol))
		b.writeFingerprint(c.censusTypeFingerprint(t.AsStringMappingType().target))
	default:
		return unkeyableFingerprint
	}
	return b.fingerprint()
}

// censusTypeParameterFingerprint keys a type parameter by its declaring symbol. Clones made for
// signature and mapped type instantiations are keyed by their origin, not by their mapper: the
// mapper maps the origin to the clone inside the instantiation that owns it, so it cannot be
// part of the clone's identity, and the owning instantiation carries the arguments that tell
// clones apart. Type parameters invented during inference have no declaration and no key.
func (c *Checker) censusTypeParameterFingerprint(t *Type) fingerprint {
	if t.symbol == nil || t.symbol.Flags&ast.SymbolFlagsTransient != 0 && len(t.symbol.Declarations) == 0 {
		return unkeyableFingerprint
	}
	d := t.AsTypeParameter()
	var b fingerprintBuilder
	b.writeByte('p')
	b.writeFingerprint(c.censusSymbolFingerprint(t.symbol))
	if d.isThisType {
		b.writeByte('t')
	}
	if d.target != nil {
		b.writeByte('c')
	}
	return b.fingerprint()
}

func (c *Checker) censusObjectTypeFingerprint(t *Type) fingerprint {
	var b fingerprintBuilder
	switch {
	case t.objectFlags&(ObjectFlagsReverseMapped|ObjectFlagsEvolvingArray) != 0:
		// Inferred from expressions.
		return unkeyableFingerprint
	case t.objectFlags&ObjectFlagsClassOrInterface != 0:
		b.writeByte('C')
		b.writeFingerprint(c.censusSymbolFingerprint(t.symbol))
	case t.objectFlags&ObjectFlagsReference != 0:
		ref := t.AsTypeReference()
		if tuple, ok := t.data.(*TupleType); ok {
			// A tuple target (class and interface targets are handled above): its element shape. The tuple object
			// flag alone does not identify one: cloneTypeReference copies it onto plain references to the target.
			b.writeByte('T')
			for _, e := range tuple.elementInfos {
				b.writeUint32(uint32(e.flags))
				if e.labeledDeclaration != nil {
					b.writeNodeIdentity(e.labeledDeclaration)
				}
			}
			if tuple.readonly {
				b.writeByte('r')
			}
			return b.fingerprint()
		}
		b.writeByte('R')
		b.writeFingerprint(c.censusTypeFingerprint(ref.target))
		if ref.resolvedTypeArguments != nil {
			b.writeTypeFingerprints(c, ref.resolvedTypeArguments)
		} else {
			// Deferred: the reference node and the mapper determine the type arguments.
			b.writeByte('d')
			b.writeNodeIdentity(ref.node)
			b.writeFingerprint(c.censusMapperFingerprint(ref.mapper))
		}
		b.writeAliasFingerprint(c, t.alias)
	case t.objectFlags&(ObjectFlagsObjectLiteral|ObjectFlagsJsxAttributes|ObjectFlagsArrayLiteral|ObjectFlagsObjectLiteralPatternWithComputedProperties|ObjectFlagsObjectRestType) != 0:
		// Anchored to an expression.
		return unkeyableFingerprint
	default:
		// Anonymous, mapped and instantiation expression types.
		obj := t.AsObjectType()
		if obj.target != nil {
			b.writeByte('A')
			b.writeFingerprint(c.censusTypeFingerprint(obj.target))
			b.writeFingerprint(c.censusMapperFingerprint(obj.mapper))
		} else {
			if t.symbol == nil || censusIsExpressionAnchored(t.symbol) {
				return unkeyableFingerprint
			}
			b.writeByte('a')
			b.writeUint32(uint32(t.objectFlags & ObjectFlagsObjectTypeKindMask))
			b.writeFingerprint(c.censusSymbolFingerprint(t.symbol))
			switch {
			case t.objectFlags&ObjectFlagsMapped != 0:
				b.writeNodeIdentity(t.AsMappedType().declaration.AsNode())
			case t.objectFlags&ObjectFlagsInstantiationExpressionType != 0:
				b.writeNodeIdentity(t.AsInstantiationExpressionType().node)
			}
		}
		b.writeAliasFingerprint(c, t.alias)
	}
	return b.fingerprint()
}

// censusIsExpressionAnchored reports whether the declaration of an anonymous type's symbol is an
// expression, i.e. the type describes a value of the checked program rather than a declaration.
func censusIsExpressionAnchored(symbol *ast.Symbol) bool {
	declaration := censusSymbolAnchor(symbol)
	return declaration != nil && (ast.IsFunctionExpressionOrArrowFunction(declaration) || ast.IsClassExpression(declaration) || ast.IsObjectLiteralExpression(declaration))
}

// censusMapperFingerprint is structural over the sources and targets of a mapper. Deferred,
// function and inference mappers compute their targets from checker state, so they have no key.
// Mappers are not memoised: each is fingerprinted about once, through the memoised type or
// symbol that holds it.
func (c *Checker) censusMapperFingerprint(m *TypeMapper) fingerprint {
	var b fingerprintBuilder
	if m == nil {
		b.writeByte(0)
		return b.fingerprint()
	}
	switch d := m.data.(type) {
	case *SimpleTypeMapper:
		b.writeByte('s')
		b.writeFingerprint(c.censusTypeFingerprint(d.source))
		b.writeFingerprint(c.censusTypeFingerprint(d.target))
	case *ArrayTypeMapper:
		b.writeByte('a')
		b.writeTypeFingerprints(c, d.sources)
		b.writeTypeFingerprints(c, d.targets)
	case *ArrayToSingleTypeMapper:
		b.writeByte('1')
		b.writeTypeFingerprints(c, d.sources)
		b.writeFingerprint(c.censusTypeFingerprint(d.target))
	case *MergedTypeMapper:
		b.writeByte('m')
		b.writeFingerprint(c.censusMapperFingerprint(d.m1))
		b.writeFingerprint(c.censusMapperFingerprint(d.m2))
	case *CompositeTypeMapper:
		b.writeByte('c')
		b.writeFingerprint(c.censusMapperFingerprint(d.m1))
		b.writeFingerprint(c.censusMapperFingerprint(d.m2))
	default:
		return unkeyableFingerprint
	}
	return b.fingerprint()
}

// censusSignatureFingerprint keys a signature by its declaration, mapper and, for instantiated
// signatures, target. Union and intersection signatures and signatures the checker synthesises
// without a declaration have no key.
func (c *Checker) censusSignatureFingerprint(sig *Signature) fingerprint {
	cs := c.census
	if int(sig.id) >= len(cs.signatureFingerprints) {
		cs.signatureFingerprints = slices.Grow(cs.signatureFingerprints, int(c.SignatureCount)+1-len(cs.signatureFingerprints))[:int(c.SignatureCount)+1]
	}
	if fp, ok := cs.signatureFingerprints[sig.id].cached(); ok {
		return fp
	}
	cs.signatureFingerprints[sig.id].state = fingerprintInProgress
	fp := c.censusSignatureFingerprintWorker(sig)
	cs.signatureFingerprints[sig.id] = fingerprintMemo{fingerprint: fp, state: fingerprintDone}
	return fp
}

func (c *Checker) censusSignatureFingerprintWorker(sig *Signature) fingerprint {
	var b fingerprintBuilder
	switch {
	case sig.id <= c.census.initSignatures:
		b.writeByte('0')
		b.writeUint32(uint32(sig.id))
	case sig.composite != nil || sig.declaration == nil:
		return unkeyableFingerprint
	default:
		b.writeByte('g')
		b.writeNodeIdentity(sig.declaration)
		b.writeFingerprint(c.censusMapperFingerprint(sig.mapper))
		if sig.target != nil {
			b.writeFingerprint(c.censusSignatureFingerprint(sig.target))
		}
	}
	return b.fingerprint()
}
