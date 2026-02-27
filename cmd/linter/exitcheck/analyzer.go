package exitcheck

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer проверяет использование panic, log.Fatal и os.Exit в коде
var Analyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "проверяет использование panic, log.Fatal и os.Exit вне функции main пакета main",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		// Пропускаем автогенерированные файлы
		if isGeneratedFile(file) {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			// Проверяем встроенную функцию panic
			if call, ok := n.(*ast.CallExpr); ok {
				checkPanic(pass, call)
				checkExitFunctions(pass, call, file)
			}

			return true
		})
	}

	return nil, nil
}

// isGeneratedFile проверяет, является ли файл автогенерированным
func isGeneratedFile(file *ast.File) bool {
	// Проверяем комментарии в начале файла
	if file.Doc != nil {
		for _, c := range file.Doc.List {
			if strings.Contains(c.Text, "Code generated") ||
				strings.Contains(c.Text, "DO NOT EDIT") {
				return true
			}
		}
	}

	// Также проверяем все комментарии в файле
	for _, commentGroup := range file.Comments {
		for _, comment := range commentGroup.List {
			if strings.Contains(comment.Text, "Code generated") ||
				strings.Contains(comment.Text, "DO NOT EDIT") {
				return true
			}
		}
	}

	return false
}

// checkPanic проверяет использование встроенной функции panic
func checkPanic(pass *analysis.Pass, call *ast.CallExpr) {
	if ident, ok := call.Fun.(*ast.Ident); ok {
		if ident.Name == "panic" {
			// Проверяем, что это действительно встроенная функция panic
			if obj := pass.TypesInfo.Uses[ident]; obj != nil {
				if builtin, ok := obj.(*types.Builtin); ok && builtin.Name() == "panic" {
					pass.Reportf(call.Pos(), "использование встроенной функции panic")
				}
			}
		}
	}
}

// checkExitFunctions проверяет вызовы log.Fatal и os.Exit вне функции main пакета main
func checkExitFunctions(pass *analysis.Pass, call *ast.CallExpr, file *ast.File) {
	// Проверяем, является ли это вызовом log.Fatal или os.Exit
	var isExitCall bool
	var functionName string

	switch fun := call.Fun.(type) {
	case *ast.SelectorExpr:
		// Проверяем вызовы вида pkg.Function()
		if ident, ok := fun.X.(*ast.Ident); ok {
			// Получаем объект для идентификатора пакета
			if obj := pass.TypesInfo.Uses[ident]; obj != nil {
				if pkgName, ok := obj.(*types.PkgName); ok {
					pkgPath := pkgName.Imported().Path()
					funcName := fun.Sel.Name

					// Проверяем log.Fatal
					if pkgPath == "log" && funcName == "Fatal" {
						isExitCall = true
						functionName = "log.Fatal"
					}

					// Проверяем os.Exit
					if pkgPath == "os" && funcName == "Exit" {
						isExitCall = true
						functionName = "os.Exit"
					}
				}
			}
		}
	}

	if !isExitCall {
		return
	}

	// Если это вызов exit-функции, проверяем, находится ли он в функции main пакета main
	if isInMainFunction(call, file) && pass.Pkg.Name() == "main" {
		return
	}

	// В противном случае сообщаем об ошибке
	pass.Reportf(call.Pos(), "вызов %s вне функции main пакета main", functionName)
}

// isInMainFunction проверяет, находится ли узел внутри функции main
func isInMainFunction(node ast.Node, file *ast.File) bool {
	// Ищем функцию, в которой находится узел
	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			if funcDecl.Name.Name == "main" {
				// Проверяем, находится ли node внутри этой функции
				if funcDecl.Body != nil && funcDecl.Pos() <= node.Pos() && node.End() <= funcDecl.End() {
					return true
				}
			}
		}
	}

	return false
}
