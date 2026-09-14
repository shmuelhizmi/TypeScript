package compiler

import (
	"sync"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/binder"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/diagnostics"
	"github.com/microsoft/TypeScript/tsc/internal/outputpaths"
	"github.com/microsoft/TypeScript/tsc/internal/printer"
	"github.com/microsoft/TypeScript/tsc/internal/sourcemap"
	"github.com/microsoft/TypeScript/tsc/internal/stringutil"
	"github.com/microsoft/TypeScript/tsc/internal/tracing"
	"github.com/microsoft/TypeScript/tsc/internal/transformers"
	"github.com/microsoft/TypeScript/tsc/internal/transformers/declarations"
	"github.com/microsoft/TypeScript/tsc/internal/transformers/estransforms"
	"github.com/microsoft/TypeScript/tsc/internal/transformers/inliners"
	"github.com/microsoft/TypeScript/tsc/internal/transformers/jsxtransforms"
	"github.com/microsoft/TypeScript/tsc/internal/transformers/moduletransforms"
	"github.com/microsoft/TypeScript/tsc/internal/transformers/tstransforms"
	"github.com/microsoft/TypeScript/tsc/internal/tsoptions"
	"github.com/microsoft/TypeScript/tsc/internal/tspath"
)

type EmitOnly byte

const (
	EmitAll EmitOnly = iota
	EmitOnlyJs
	EmitOnlyDts
	EmitOnlyBuilderSignature
)

type emitter struct {
	host               EmitHost
	emitOnly           EmitOnly
	emitterDiagnostics ast.DiagnosticsCollection
	writer             printer.EmitTextWriter
	paths              *outputpaths.OutputPaths
	sourceFile         *ast.SourceFile
	emitResult         EmitResult
	forceEmit          bool
	writeFile          func(fileName string, text string, data *WriteFileData) error
	tr                 *tracing.Tracing
	singleThreaded     bool
}

func (e *emitter) emit() {
	if e.tr != nil {
		defer e.tr.Push(tracing.PhaseEmit, "emit", map[string]any{"path": string(e.sourceFile.Path())}, true)()
	}
	if e.emitsConcurrently() {
		e.emitConcurrently()
	} else {
		e.emitJSFile(e.sourceFile, e.paths.JsFilePath(), e.paths.SourceMapFilePath())
		e.emitDeclarationFile(e.sourceFile, e.paths.DeclarationFilePath(), e.paths.DeclarationMapPath())
	}
	e.emitResult.Diagnostics = e.emitterDiagnostics.GetDiagnostics()
}

// concurrentEmitMinLength is the source length from which a file's JavaScript is emitted concurrently with its
// declaration transforms. The two emits of such a file take long enough that, one after the other, they are the
// longest part of the program's emit phase; for smaller files the second emit context kept alive by the overlap
// would cost more memory than the time it saves.
const concurrentEmitMinLength = 256 << 10

func (e *emitter) emitsConcurrently() bool {
	return !e.singleThreaded && e.emitOnly == EmitAll && len(e.paths.JsFilePath()) != 0 && len(e.paths.DeclarationFilePath()) != 0 &&
		len(e.sourceFile.Text()) >= concurrentEmitMinLength
}

// emitConcurrently emits the JavaScript file on a second emitter and goroutine while the declaration transforms
// run. The JavaScript emit only queries the checker, each query under the resolver's lock, so the checker performs
// the declaration transforms' operations in the same order as in the sequential emit. The JavaScript emitter's
// results are folded in before the declaration file is printed, so the emitted files, source maps and diagnostics
// are recorded in the sequential emit's order.
func (e *emitter) emitConcurrently() {
	js := &emitter{
		host:       e.host,
		emitOnly:   e.emitOnly,
		writer:     printer.NewTextWriter(e.host.Options().NewLine.GetNewLineCharacter(), 0),
		paths:      e.paths,
		sourceFile: e.sourceFile,
		forceEmit:  e.forceEmit,
		writeFile:  e.writeFile,
		tr:         e.tr,
	}
	var wg sync.WaitGroup
	wg.Go(func() {
		js.emitJSFile(js.sourceFile, js.paths.JsFilePath(), js.paths.SourceMapFilePath())
	})
	if e.tr != nil {
		defer e.tr.Push(tracing.PhaseEmit, "emitDeclarationFileOrBundle", map[string]any{"declarationFilePath": e.paths.DeclarationFilePath()}, true)()
	}
	declaration, ok := e.transformDeclarationFile(e.sourceFile, e.paths.DeclarationFilePath(), e.paths.DeclarationMapPath())
	wg.Wait()
	e.emitResult.EmittedFiles = js.emitResult.EmittedFiles
	e.emitResult.SourceMaps = js.emitResult.SourceMaps
	if js.emitResult.EmitSkipped {
		e.emitResult.EmitSkipped = true
	}
	for _, diagnostic := range js.emitterDiagnostics.GetDiagnostics() {
		e.emitterDiagnostics.Add(diagnostic)
	}
	if ok {
		e.printDeclarationFile(declaration)
	}
}

type declarationTransformer interface {
	TransformSourceFile(sourceFile *ast.SourceFile) *ast.SourceFile
	GetDiagnostics() []*ast.Diagnostic
}

func (e *emitter) getDeclarationTransformers(emitContext *printer.EmitContext, sourceFile *ast.SourceFile, declarationFilePath string, declarationMapPath string) []declarationTransformer {
	forceDtsEmit := e.emitOnly == EmitOnlyBuilderSignature || e.forceEmit && e.emitOnly == EmitOnlyDts
	return []declarationTransformer{
		declarations.NewDeclarationTransformer(e.host, emitContext, e.host.Options(), declarationFilePath, declarationMapPath),
		declarations.NewSupplementalReferencesTransformer(e.host, sourceFile, declarationFilePath, forceDtsEmit),
	}
}

func (e *emitter) runScriptTransformers(emitContext *printer.EmitContext, sourceFile *ast.SourceFile) *ast.SourceFile {
	if e.tr != nil {
		defer e.tr.Push(tracing.PhaseEmit, "transformNodes", map[string]any{"path": string(sourceFile.Path())}, false)()
	}
	for _, transformer := range getScriptTransformers(emitContext, e.host, sourceFile) {
		sourceFile = transformer.TransformSourceFile(sourceFile)
	}
	return sourceFile
}

func (e *emitter) runDeclarationTransformers(emitContext *printer.EmitContext, sourceFile *ast.SourceFile, declarationFilePath, declarationMapPath string) (*ast.SourceFile, []*ast.Diagnostic) {
	if e.tr != nil {
		defer e.tr.Push(tracing.PhaseEmit, "transformNodes", map[string]any{"path": string(sourceFile.Path())}, false)()
	}
	var diags []*ast.Diagnostic
	for _, transformer := range e.getDeclarationTransformers(emitContext, sourceFile, declarationFilePath, declarationMapPath) {
		sourceFile = transformer.TransformSourceFile(sourceFile)
		diags = append(diags, transformer.GetDiagnostics()...)
	}
	return sourceFile, diags
}

func getModuleTransformer(opts *transformers.TransformOptions) *transformers.Transformer {
	switch opts.CompilerOptions.GetEmitModuleKind() {
	case core.ModuleKindPreserve:
		// `ESModuleTransformer` contains logic for preserving CJS input syntax in `--module preserve`
		return moduletransforms.NewESModuleTransformer(opts)

	case core.ModuleKindESNext,
		core.ModuleKindES2022,
		core.ModuleKindES2020,
		core.ModuleKindES2015,
		core.ModuleKindNode20,
		core.ModuleKindNode18,
		core.ModuleKindNode16,
		core.ModuleKindNodeNext,
		core.ModuleKindCommonJS:
		return moduletransforms.NewImpliedModuleTransformer(opts)

	default:
		return moduletransforms.NewCommonJSModuleTransformer(opts)
	}
}

func getScriptTransformers(emitContext *printer.EmitContext, host printer.EmitHost, sourceFile *ast.SourceFile) []*transformers.Transformer {
	var tx []*transformers.Transformer
	options := host.Options()

	// JS files don't use reference calculations as they don't do import elision, no need to calculate it
	importElisionEnabled := !options.VerbatimModuleSyntax.IsTrue() && !ast.IsInJSFile(sourceFile.AsNode())
	jsxTransformEnabled := options.GetJSXTransformEnabled() && sourceFile.LanguageVariant == core.LanguageVariantJSX

	emitResolver := host.GetEmitResolver()

	var referenceResolver binder.ReferenceResolver
	if importElisionEnabled || jsxTransformEnabled || !options.GetIsolatedModules() || options.EmitDecoratorMetadata.IsTrue() {
		referenceResolver = emitResolver
	} else {
		referenceResolver = binder.NewReferenceResolver(options, binder.ReferenceResolverHooks{})
	}

	opts := transformers.TransformOptions{
		Context:                   emitContext,
		CompilerOptions:           options,
		Resolver:                  referenceResolver,
		EmitResolver:              emitResolver,
		GetEmitModuleFormatOfFile: host.GetEmitModuleFormatOfFile,
	}

	// transform TypeScript syntax
	{
		// use type nodes to add metadata decorators
		if options.EmitDecoratorMetadata.IsTrue() {
			tx = append(tx, tstransforms.NewMetadataTransformer(&opts))
		}

		// erase types
		tx = append(tx, tstransforms.NewTypeEraserTransformer(&opts))

		// elide imports
		if importElisionEnabled {
			tx = append(tx, tstransforms.NewImportElisionTransformer(&opts))
		}

		// transform `enum`, `namespace`, and parameter properties
		tx = append(tx, tstransforms.NewRuntimeSyntaxTransformer(&opts))

		if options.ExperimentalDecorators.IsTrue() {
			tx = append(tx, tstransforms.NewLegacyDecoratorsTransformer(&opts))
		}
	}

	if jsxTransformEnabled {
		tx = append(tx, jsxtransforms.NewJSXTransformer(&opts))
	}

	downleveler := estransforms.GetESTransformer(&opts)
	if downleveler != nil {
		tx = append(tx, downleveler)
	}

	tx = append(tx, estransforms.NewUseStrictTransformer(&opts))

	// transform module syntax
	tx = append(tx, getModuleTransformer(&opts))

	// inlining (formerly done via substitutions)
	if !options.GetIsolatedModules() {
		tx = append(tx, inliners.NewConstEnumInliningTransformer(&opts))
	}
	return tx
}

func (e *emitter) emitJSFile(sourceFile *ast.SourceFile, jsFilePath string, sourceMapFilePath string) {
	options := e.host.Options()

	if sourceFile == nil || e.emitOnly != EmitAll && e.emitOnly != EmitOnlyJs || len(jsFilePath) == 0 {
		return
	}

	if !e.forceEmit && (options.NoEmit == core.TSTrue || e.host.IsEmitBlocked(jsFilePath)) {
		e.emitResult.EmitSkipped = true
		return
	}

	if e.tr != nil {
		defer e.tr.Push(tracing.PhaseEmit, "emitJsFileOrBundle", map[string]any{"jsFilePath": jsFilePath}, true)()
	}

	emitContext, putEmitContext := printer.GetEmitContext()
	defer putEmitContext()

	sourceFile = e.runScriptTransformers(emitContext, sourceFile)

	printerOptions := printer.PrinterOptions{
		RemoveComments:  options.RemoveComments.IsTrue(),
		NewLine:         options.NewLine,
		NoEmitHelpers:   options.NoEmitHelpers.IsTrue(),
		SourceMap:       options.SourceMap.IsTrue(),
		InlineSourceMap: options.InlineSourceMap.IsTrue(),
		InlineSources:   options.InlineSources.IsTrue(),
		Target:          options.Target,
		// !!!
	}

	// create a printer to print the nodes
	printer := printer.NewPrinter(printerOptions, printer.PrintHandlers{
		// !!!
	}, emitContext)

	e.printSourceFile(jsFilePath, sourceMapFilePath, sourceFile, printer, options, shouldEmitSourceMaps(options, sourceFile))
}

func (e *emitter) emitDeclarationFile(sourceFile *ast.SourceFile, declarationFilePath string, declarationMapPath string) {
	if sourceFile == nil || e.emitOnly == EmitOnlyJs || len(declarationFilePath) == 0 {
		return
	}

	if e.tr != nil {
		defer e.tr.Push(tracing.PhaseEmit, "emitDeclarationFileOrBundle", map[string]any{"declarationFilePath": declarationFilePath}, true)()
	}

	if declaration, ok := e.transformDeclarationFile(sourceFile, declarationFilePath, declarationMapPath); ok {
		e.printDeclarationFile(declaration)
	}
}

// transformedDeclarationFile is a source file after its declaration transforms, ready to be printed.
type transformedDeclarationFile struct {
	sourceFile          *ast.SourceFile
	contentMappedSource *ast.SourceFile
	declarationFilePath string
	declarationMapPath  string
	emitDeclarationMap  bool
	emitContext         *printer.EmitContext
	putEmitContext      func()
}

// transformDeclarationFile runs the declaration transforms and records their diagnostics; it reports false when
// the declaration file is not to be printed.
func (e *emitter) transformDeclarationFile(sourceFile *ast.SourceFile, declarationFilePath string, declarationMapPath string) (transformedDeclarationFile, bool) {
	options := e.host.Options()
	emitDeclarationMap := e.emitOnly != EmitOnlyBuilderSignature && options.DeclarationMap.IsTrue()

	emitContext, putEmitContext := printer.GetEmitContext()
	transformed, diags := e.runDeclarationTransformers(emitContext, sourceFile, declarationFilePath, declarationMapPath)

	for _, elem := range diags {
		// Add declaration transform diagnostics to emit diagnostics
		e.emitterDiagnostics.Add(elem)
	}

	if !e.forceEmit && e.emitOnly != EmitOnlyBuilderSignature && (options.NoEmit == core.TSTrue || e.host.IsEmitBlocked(declarationFilePath)) {
		e.emitResult.EmitSkipped = true
		putEmitContext()
		return transformedDeclarationFile{}, false
	}

	declBlocked := len(diags) > 0 && !e.forceEmit && e.emitOnly != EmitOnlyBuilderSignature
	if declBlocked {
		e.emitResult.EmitSkipped = true
		putEmitContext()
		return transformedDeclarationFile{}, false
	}

	return transformedDeclarationFile{
		sourceFile:          transformed,
		contentMappedSource: sourceFile,
		declarationFilePath: declarationFilePath,
		declarationMapPath:  declarationMapPath,
		emitDeclarationMap:  emitDeclarationMap,
		emitContext:         emitContext,
		putEmitContext:      putEmitContext,
	}, true
}

func (e *emitter) printDeclarationFile(declaration transformedDeclarationFile) {
	defer declaration.putEmitContext()
	options := e.host.Options()

	printerOptions := printer.PrinterOptions{
		RemoveComments: options.RemoveComments.IsTrue(),
		NewLine:        options.NewLine,
		NoEmitHelpers:  true,
		// Module: 			   options.Module, // NYI
		// ModuleResolution:   options.ModuleResolution, // NYI
		Target:          options.GetEmitScriptTarget(),
		SourceMap:       declaration.emitDeclarationMap,
		InlineSourceMap: options.InlineSourceMap.IsTrue(),
		// InlineSources:       options.InlineSources.IsTrue(), // ignored, per strada
		// ExtendedDiagnostics: options.ExtendedDiagnostics.IsTrue(), // NYI
		OnlyPrintJSDocStyle:         true,
		OmitBraceSourceMapPositions: true,
	}

	// create a printer to print the nodes
	printHandlers := printer.PrintHandlers{}
	contentMappedSource := declaration.contentMappedSource
	if spanMap := contentMappedSource.SpanMap(); declaration.emitDeclarationMap && spanMap != nil {
		originalSource := newDeclarationMapSource(contentMappedSource)
		printHandlers.MapSourcePosition = func(source sourcemap.Source, pos int) (sourcemap.Source, int, bool) {
			if source.FileName() != contentMappedSource.FileName() {
				return source, pos, true
			}
			mapped, ok := spanMap.VirtualToOriginalPositionExact(core.TextPos(pos))
			if !ok {
				return nil, 0, false
			}
			return originalSource, int(mapped), true
		}
	}
	printer := printer.NewPrinter(printerOptions, printHandlers, declaration.emitContext)

	declarationMapOptions := &core.CompilerOptions{
		SourceMap:  core.IfElse(declaration.emitDeclarationMap, core.TSTrue, core.TSFalse),
		SourceRoot: options.SourceRoot,
		MapRoot:    options.MapRoot,
		// Explicitly do not pass through either inline option.
	}
	e.printSourceFile(declaration.declarationFilePath, declaration.declarationMapPath, declaration.sourceFile, printer, declarationMapOptions, shouldEmitSourceMaps(declarationMapOptions, declaration.sourceFile))
}

type declarationMapSource struct {
	fileName string
	text     string
	lineMap  []core.TextPos
}

func newDeclarationMapSource(sourceFile *ast.SourceFile) *declarationMapSource {
	text := sourceFile.OriginalText()
	return &declarationMapSource{
		fileName: sourceFile.OriginalFileName(),
		text:     text,
		lineMap:  []core.TextPos(core.ComputeECMALineStarts(text)),
	}
}

func (s *declarationMapSource) FileName() string            { return s.fileName }
func (s *declarationMapSource) Text() string                { return s.text }
func (s *declarationMapSource) ECMALineMap() []core.TextPos { return s.lineMap }

func (e *emitter) printSourceFile(jsFilePath string, sourceMapFilePath string, sourceFile *ast.SourceFile, printer_ *printer.Printer, mapOptions *core.CompilerOptions, shouldEmitSourceMaps bool) {
	// !!! sourceMapGenerator
	options := e.host.Options()
	var sourceMapGenerator *sourcemap.Generator
	if shouldEmitSourceMaps {
		sourceMapGenerator = sourcemap.NewGenerator(
			tspath.GetBaseFileName(tspath.NormalizeSlashes(jsFilePath)),
			getSourceRoot(mapOptions),
			e.getSourceMapDirectory(mapOptions, jsFilePath, sourceFile),
			tspath.ComparePathsOptions{
				UseCaseSensitiveFileNames: e.host.UseCaseSensitiveFileNames(),
				CurrentDirectory:          e.host.GetCurrentDirectory(),
			},
		)
	}

	printer_.Write(sourceFile.AsNode(), sourceFile, e.writer, sourceMapGenerator)

	sourceMapUrlPos := -1
	if sourceMapGenerator != nil {
		if mapOptions.SourceMap.IsTrue() || mapOptions.InlineSourceMap.IsTrue() {
			e.emitResult.SourceMaps = append(e.emitResult.SourceMaps, &SourceMapEmitResult{
				InputSourceFileNames: sourceMapGenerator.Sources(),
				SourceMap:            sourceMapGenerator.RawSourceMap(),
				GeneratedFile:        jsFilePath,
			})
		}

		sourceMappingURL := e.getSourceMappingURL(
			mapOptions,
			sourceMapGenerator,
			jsFilePath,
			sourceMapFilePath,
			sourceFile,
		)

		if len(sourceMappingURL) > 0 {
			if !e.writer.IsAtStartOfLine() {
				e.writer.RawWrite(core.IfElse(options.NewLine == core.NewLineKindCRLF, "\r\n", "\n"))
			}
			sourceMapUrlPos = e.writer.GetTextPos()
			e.writer.WriteComment("//# sourceMappingURL=")
			e.writer.WriteComment(sourceMappingURL)
		}

		// Write the source map
		if len(sourceMapFilePath) > 0 {
			sourceMap := sourceMapGenerator.String()
			err := e.writeText(sourceMapFilePath, sourceMap, &WriteFileData{SourceFile: e.sourceFile})
			if err != nil {
				e.emitterDiagnostics.Add(ast.NewCompilerDiagnostic(diagnostics.Could_not_write_file_0_Colon_1, jsFilePath, err.Error()))
			} else {
				e.emitResult.EmittedFiles = append(e.emitResult.EmittedFiles, sourceMapFilePath)
			}
		}
	} else {
		e.writer.WriteLine()
	}

	// Write the output file
	text := e.writer.String()
	if options.EmitBOM.IsTrue() {
		text = stringutil.AddUTF8ByteOrderMark(text)
	}
	data := &WriteFileData{
		SourceMapUrlPos: sourceMapUrlPos,
		Diagnostics:     e.emitterDiagnostics.GetDiagnostics(),
		SourceFile:      e.sourceFile,
	}
	err := e.writeText(jsFilePath, text, data)
	skippedDtsWrite := data.SkippedDtsWrite
	if err != nil {
		e.emitterDiagnostics.Add(ast.NewCompilerDiagnostic(diagnostics.Could_not_write_file_0_Colon_1, jsFilePath, err.Error()))
	} else if !skippedDtsWrite {
		e.emitResult.EmittedFiles = append(e.emitResult.EmittedFiles, jsFilePath)
	}

	// Reset state
	e.writer.Clear()
}

func (e *emitter) writeText(fileName string, text string, data *WriteFileData) error {
	if e.writeFile != nil {
		return e.writeFile(fileName, text, data)
	}
	return e.host.WriteFile(fileName, text)
}

func shouldEmitSourceMaps(mapOptions *core.CompilerOptions, sourceFile *ast.SourceFile) bool {
	return (mapOptions.SourceMap.IsTrue() || mapOptions.InlineSourceMap.IsTrue()) &&
		!tspath.FileExtensionIs(sourceFile.FileName(), tspath.ExtensionJson)
}

func getSourceRoot(mapOptions *core.CompilerOptions) string {
	// Normalize source root and make sure it has trailing "/" so that it can be used to combine paths with the
	// relative paths of the sources list in the sourcemap
	sourceRoot := tspath.NormalizeSlashes(mapOptions.SourceRoot)
	if len(sourceRoot) > 0 {
		sourceRoot = tspath.EnsureTrailingDirectorySeparator(sourceRoot)
	}
	return sourceRoot
}

func (e *emitter) getSourceMapDirectory(mapOptions *core.CompilerOptions, filePath string, sourceFile *ast.SourceFile) string {
	if len(mapOptions.SourceRoot) > 0 {
		return e.host.CommonSourceDirectory()
	}
	if len(mapOptions.MapRoot) > 0 {
		sourceMapDir := tspath.NormalizeSlashes(mapOptions.MapRoot)
		if sourceFile != nil {
			// For modules or multiple emit files the mapRoot will have directory structure like the sources
			// So if src\a.ts and src\lib\b.ts are compiled together user would be moving the maps into mapRoot\a.js.map and mapRoot\lib\b.js.map
			sourceMapDir = tspath.GetDirectoryPath(outputpaths.GetSourceFilePathInNewDir(
				sourceFile.FileName(),
				sourceMapDir,
				e.host.GetCurrentDirectory(),
				e.host.CommonSourceDirectory(),
				e.host.UseCaseSensitiveFileNames(),
			))
		}
		if tspath.GetRootLength(sourceMapDir) == 0 {
			// The relative paths are relative to the common directory
			sourceMapDir = tspath.CombinePaths(e.host.CommonSourceDirectory(), sourceMapDir)
		}
		return sourceMapDir
	}
	return tspath.GetDirectoryPath(tspath.NormalizePath(filePath))
}

func (e *emitter) getSourceMappingURL(mapOptions *core.CompilerOptions, sourceMapGenerator *sourcemap.Generator, filePath string, sourceMapFilePath string, sourceFile *ast.SourceFile) string {
	if mapOptions.InlineSourceMap.IsTrue() {
		// Encode the sourceMap into the sourceMap url
		return sourceMapGenerator.Base64DataURL()
	}

	sourceMapFile := tspath.GetBaseFileName(tspath.NormalizeSlashes(sourceMapFilePath))
	if len(mapOptions.MapRoot) > 0 {
		sourceMapDir := tspath.NormalizeSlashes(mapOptions.MapRoot)
		if sourceFile != nil {
			// For modules or multiple emit files the mapRoot will have directory structure like the sources
			// So if src\a.ts and src\lib\b.ts are compiled together user would be moving the maps into mapRoot\a.js.map and mapRoot\lib\b.js.map
			sourceMapDir = tspath.GetDirectoryPath(outputpaths.GetSourceFilePathInNewDir(
				sourceFile.FileName(),
				sourceMapDir,
				e.host.GetCurrentDirectory(),
				e.host.CommonSourceDirectory(),
				e.host.UseCaseSensitiveFileNames(),
			))
		}
		if tspath.GetRootLength(sourceMapDir) == 0 {
			// The relative paths are relative to the common directory
			sourceMapDir = tspath.CombinePaths(e.host.CommonSourceDirectory(), sourceMapDir)
			return stringutil.EncodeURI(
				tspath.GetRelativePathToDirectoryOrUrl(
					tspath.GetDirectoryPath(tspath.NormalizePath(filePath)), // get the relative sourceMapDir path based on jsFilePath
					tspath.CombinePaths(sourceMapDir, sourceMapFile),        // this is where user expects to see sourceMap
					/*isAbsolutePathAnUrl*/ true,
					tspath.ComparePathsOptions{
						UseCaseSensitiveFileNames: e.host.UseCaseSensitiveFileNames(),
						CurrentDirectory:          e.host.GetCurrentDirectory(),
					},
				),
			)
		} else {
			return stringutil.EncodeURI(tspath.CombinePaths(sourceMapDir, sourceMapFile))
		}
	}
	return stringutil.EncodeURI(sourceMapFile)
}

type SourceFileMayBeEmittedHost interface {
	Options() *core.CompilerOptions
	GetProjectReferenceFromSource(path tspath.Path) *tsoptions.SourceOutputAndProjectReference
	IsSourceFileFromExternalLibrary(file *ast.SourceFile) bool
	GetCurrentDirectory() string
	UseCaseSensitiveFileNames() bool
	SourceFiles() []*ast.SourceFile
}

func sourceFileMayBeEmitted(sourceFile *ast.SourceFile, host SourceFileMayBeEmittedHost, forceDtsEmit bool, forceJsEmit bool) bool {
	// TODO: move this to outputpaths?
	options := host.Options()
	// Js files are emitted only if option is enabled
	if !forceJsEmit && options.NoEmitForJsFiles.IsTrue() && ast.IsSourceFileJS(sourceFile) {
		return false
	}

	// Declaration files are not emitted
	if sourceFile.IsDeclarationFile {
		return false
	}

	// Runtime output for content-mapped files is owned by the external content mapper or build tool. Only
	// include them in the emit set when their transformed TypeScript can produce declarations.
	if sourceFile.ContentMapper() != "" && !forceDtsEmit && !options.GetEmitDeclarations() {
		return false
	}

	// Source file from node_modules are not emitted
	if host.IsSourceFileFromExternalLibrary(sourceFile) {
		return false
	}

	// forcing dts emit => file needs to be emitted
	if forceDtsEmit || forceJsEmit {
		return true
	}

	// Check other conditions for file emit
	// Source files from referenced projects are not emitted
	if host.GetProjectReferenceFromSource(sourceFile.Path()) != nil {
		return false
	}

	// Any non json file should be emitted
	if !ast.IsJsonSourceFile(sourceFile) {
		return true
	}

	// Json file is not emitted if outDir is not specified
	if options.OutDir == "" {
		return false
	}

	// Otherwise, if rootDir is specified or a config file exists, we know the common source directory and can check if the file would be emitted in the same location
	if options.RootDir != "" || options.ConfigFilePath != "" {
		commonDir := tspath.GetNormalizedAbsolutePath(outputpaths.GetCommonSourceDirectory(options, func() []string { return nil }, host.GetCurrentDirectory(), host.UseCaseSensitiveFileNames(), nil), host.GetCurrentDirectory())
		outputPath := outputpaths.GetSourceFilePathInNewDirWorker(sourceFile.FileName(), options.OutDir, host.GetCurrentDirectory(), commonDir, host.UseCaseSensitiveFileNames())
		if tspath.ComparePaths(sourceFile.FileName(), outputPath, tspath.ComparePathsOptions{
			UseCaseSensitiveFileNames: host.UseCaseSensitiveFileNames(),
			CurrentDirectory:          host.GetCurrentDirectory(),
		}) == 0 {
			return false
		}
	}

	return true
}

func getSourceFilesToEmit(host SourceFileMayBeEmittedHost, targetSourceFiles []*ast.SourceFile, forceDtsEmit bool, forceJsEmit bool) []*ast.SourceFile {
	if targetSourceFiles == nil {
		targetSourceFiles = host.SourceFiles()
	}
	return core.Filter(targetSourceFiles, func(sourceFile *ast.SourceFile) bool {
		return sourceFileMayBeEmitted(sourceFile, host, forceDtsEmit, forceJsEmit)
	})
}

func isSourceFileNotJson(file *ast.SourceFile) bool {
	return !ast.IsJsonSourceFile(file)
}

func getDeclarationDiagnostics(host EmitHost, file *ast.SourceFile) []*ast.Diagnostic {
	// TODO: use p.getSourceFilesToEmit cache
	fullFiles := core.Filter(getSourceFilesToEmit(host, core.SingleElementSlice(file), false, false), isSourceFileNotJson)
	if !core.Some(fullFiles, func(f *ast.SourceFile) bool { return f == file }) {
		return []*ast.Diagnostic{}
	}
	options := host.Options()
	transform := declarations.NewDeclarationTransformer(host, nil, options, "", "")
	transform.TransformSourceFile(file)
	return transform.GetDiagnostics()
}
