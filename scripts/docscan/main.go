package main

import (
    "bufio"
    "bytes"
    "flag"
    "fmt"
    "go/ast"
    "go/parser"
    "go/token"
    "io/fs"
    "os"
    "path/filepath"
    "regexp"
    "sort"
    "strings"
)

// Tool to scan exported functions/methods and compare with docs/*.md
// Usage: go run scripts/docscan/main.go

type Symbol struct {
    Pkg   string
    Recv  string // receiver type for methods
    Name  string
    Sig   string // simplified signature
    File  string
}

func (s Symbol) Key() string {
    if s.Recv != "" {
        return fmt.Sprintf("%s.(%s).%s%s", s.Pkg, s.Recv, s.Name, s.Sig)
    }
    return fmt.Sprintf("%s.%s%s", s.Pkg, s.Name, s.Sig)
}

func (s Symbol) Short() string {
    if s.Recv != "" {
        return fmt.Sprintf("(%s).%s%s", s.Recv, s.Name, s.Sig)
    }
    return fmt.Sprintf("%s%s", s.Name, s.Sig)
}

func main() {
    var root string
    flag.StringVar(&root, "root", ".", "project root")
    flag.Parse()

    exports, err := collectExports(root)
    if err != nil {
        fmt.Fprintf(os.Stderr, "collect error: %v\n", err)
        os.Exit(1)
    }
    docs, err := collectDocs(filepath.Join(root, "docs"))
    if err != nil {
        fmt.Fprintf(os.Stderr, "docs error: %v\n", err)
        os.Exit(1)
    }

    // Map docs by package inferred from filename like docs/zstring.md -> zstring
    // docs map: pkg -> set of simplified prototypes (function or method short forms)

    missingDocs := make(map[string][]Symbol)
    undocumented := 0
    for _, s := range exports {
        set := docs[s.Pkg]
        short := s.Short()
        if _, ok := set[strings.TrimSpace(short)]; !ok {
            missingDocs[s.Pkg] = append(missingDocs[s.Pkg], s)
            undocumented++
        }
    }

    // Find doc-only entries that don't match any export
    extras := make(map[string][]string)
    exportSet := make(map[string]struct{})
    for _, s := range exports {
        exportSet[s.Pkg+"::"+s.Short()] = struct{}{}
    }
    for pkg, set := range docs {
        for item := range set {
            if _, ok := exportSet[pkg+"::"+item]; !ok {
                extras[pkg] = append(extras[pkg], item)
            }
        }
    }

    packages := uniqPkgs(exports)
    sort.Strings(packages)

    if len(packages) == 0 {
        fmt.Println("No Go packages found.")
        return
    }

    fmt.Println("Docscan Report")
    fmt.Println("================")
    for _, pkg := range packages {
        fmt.Printf("Package: %s\n", pkg)
        // Missing docs
        if m := missingDocs[pkg]; len(m) > 0 {
            fmt.Println("  Missing docs for:")
            sort.Slice(m, func(i, j int) bool { return m[i].Short() < m[j].Short() })
            for _, s := range m {
                fmt.Printf("  - %s\n", s.Short())
            }
        } else {
            fmt.Println("  Missing docs for: none")
        }

        // Extras
        if e := extras[pkg]; len(e) > 0 {
            sort.Strings(e)
            fmt.Println("  Docs mention non-existent:")
            for _, x := range e {
                fmt.Printf("  - %s\n", x)
            }
        } else {
            fmt.Println("  Docs mention non-existent: none")
        }
        fmt.Println()
    }
}

func uniqPkgs(syms []Symbol) []string {
    m := map[string]struct{}{}
    for _, s := range syms {
        m[s.Pkg] = struct{}{}
    }
    out := make([]string, 0, len(m))
    for k := range m {
        out = append(out, k)
    }
    return out
}

func collectExports(root string) ([]Symbol, error) {
    var syms []Symbol
    fset := token.NewFileSet()
    // Iterate top-level subpackages only (e.g., zstring, ztime...) skipping vendor and docs, node_modules
    filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return nil
        }
        if d.IsDir() {
            base := filepath.Base(path)
            switch base {
            case ".git", "node_modules", "docs", "scripts", ".history", ".jj", ".cursor", ".serena", ".kiro", ".context-forge", ".claude":
                return filepath.SkipDir
            }
            return nil
        }
        if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
            return nil
        }
        // Parse file
        file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
        if err != nil {
            return nil
        }
        pkg := file.Name.Name
        // Only consider subpackages that match directories under root (e.g., zstring)
        // Skip main or root package zlsgo (we only compare modules with docs)
        if pkg == "main" {
            return nil
        }
        // collect exported funcs and methods
        for _, decl := range file.Decls {
            fn, ok := decl.(*ast.FuncDecl)
            if !ok {
                continue
            }
            if !fn.Name.IsExported() {
                continue
            }
            recv := ""
            if fn.Recv != nil && len(fn.Recv.List) > 0 {
                // receiver type
                recv = typeExpr(fn.Recv.List[0].Type)
                recv = strings.TrimPrefix(recv, "*")
                // skip methods on unexported receiver types
                if recv != "" {
                    r := rune(recv[0])
                    if !(r >= 'A' && r <= 'Z') {
                        continue
                    }
                }
            }
            sig := buildSignature(fn.Type)
            syms = append(syms, Symbol{
                Pkg:  pkg,
                Recv: recv,
                Name: fn.Name.Name,
                Sig:  sig,
                File: path,
            })
        }
        return nil
    })
    return unique(syms), nil
}

func unique(in []Symbol) []Symbol {
    m := map[string]Symbol{}
    for _, s := range in {
        m[s.Key()] = s
    }
    out := make([]Symbol, 0, len(m))
    for _, s := range m {
        out = append(out, s)
    }
    return out
}

func buildSignature(ft *ast.FuncType) string {
    // produce simplified signature like (int, string) (string, error)
    var b bytes.Buffer
    b.WriteByte('(')
    if ft.Params != nil {
        first := true
        for _, f := range ft.Params.List {
            // repeat type for grouped names like: a, b string -> string, string
            repeat := 1
            if len(f.Names) > 0 {
                repeat = len(f.Names)
            }
            for i := 0; i < repeat; i++ {
                if !first {
                    b.WriteString(", ")
                }
                first = false
                b.WriteString(typeExpr(f.Type))
            }
        }
    }
    b.WriteByte(')')
    if ft.Results != nil && len(ft.Results.List) > 0 {
        b.WriteByte(' ')
        if len(ft.Results.List) == 1 && len(ft.Results.List[0].Names) == 0 {
            // single unnamed result, keep short form: " T"
            b.WriteString(typeExpr(ft.Results.List[0].Type))
        } else {
            // multiple results or named result(s), normalize to explicit tuple with repeated grouped types
            b.WriteByte('(')
            first := true
            for _, r := range ft.Results.List {
                repeat := 1
                if len(r.Names) > 0 {
                    repeat = len(r.Names)
                }
                for i := 0; i < repeat; i++ {
                    if !first {
                        b.WriteString(", ")
                    }
                    first = false
                    b.WriteString(typeExpr(r.Type))
                }
            }
            b.WriteByte(')')
        }
    }
    return b.String()
}

func typeExpr(e ast.Expr) string {
    switch t := e.(type) {
    case *ast.Ident:
        return t.Name
    case *ast.SelectorExpr:
        return typeExpr(t.X) + "." + t.Sel.Name
    case *ast.StarExpr:
        return "*" + typeExpr(t.X)
    case *ast.IndexExpr:
        // Generic type like Foo[T]; keep base identifier only for doc matching
        return typeExpr(t.X)
    case *ast.IndexListExpr:
        // Generic type like Foo[K, V]; keep base identifier only
        return typeExpr(t.X)
    case *ast.ArrayType:
        return "[]" + typeExpr(t.Elt)
    case *ast.MapType:
        return fmt.Sprintf("map[%s]%s", typeExpr(t.Key), typeExpr(t.Value))
    case *ast.FuncType:
        return "func" + buildSignature(t)
    case *ast.InterfaceType:
        return "interface{}"
    case *ast.StructType:
        return "struct{...}"
    case *ast.Ellipsis:
        return "..." + typeExpr(t.Elt)
    case *ast.ChanType:
        return "chan " + typeExpr(t.Value)
    case *ast.ParenExpr:
        return typeExpr(t.X)
    default:
        return fmt.Sprintf("%T", e)
    }
}

var codeBlock = regexp.MustCompile("(?s)```[a-zA-Z]*\\n(.*?)```")
var sigLine = regexp.MustCompile(`(?m)^\s*(?:func\s+)?(?:\(\s*[^)]+\)\s*)?[A-Z][A-Za-z0-9_]*\s*\(.*?\)\s*(?:\([^)]*\)|[A-Za-z0-9_\*\[\]\{\}\.\s,]+)?\s*$`)

func collectDocs(dir string) (map[string]map[string]struct{}, error) {
    out := map[string]map[string]struct{}{}
    entries, err := os.ReadDir(dir)
    if err != nil {
        if os.IsNotExist(err) {
            return out, nil
        }
        return nil, err
    }
    for _, e := range entries {
        name := e.Name()
        if e.IsDir() || !strings.HasSuffix(name, ".md") {
            continue
        }
        pkg := strings.TrimSuffix(name, ".md")
        data, err := os.ReadFile(filepath.Join(dir, name))
        if err != nil {
            continue
        }
        // extract function signatures from fenced code blocks and lines starting with func/Name(
        items := map[string]struct{}{}
        add := func(line string) {
            line = strings.TrimSpace(line)
            if line == "" {
                return
            }
            // normalize to simplified form similar to Short()
            short := simplifyDocSig(line)
            if short != "" {
                // filter only exported symbols per request
                if keepExportedOnly(short) {
                    items[short] = struct{}{}
                }
            }
        }
        // scan fenced code blocks
        for _, m := range codeBlock.FindAllSubmatch(data, -1) {
            scanner := bufio.NewScanner(bytes.NewReader(m[1]))
            for scanner.Scan() {
                t := scanner.Text()
                line := strings.TrimSpace(t)
                if strings.HasPrefix(line, "package ") || strings.HasPrefix(line, "import ") || strings.HasPrefix(line, "}") {
                    continue
                }
                if strings.HasPrefix(line, "func ") || startsWithExportedName(line) || strings.HasPrefix(line, "(") {
                    add(t)
                }
            }
        }
        // also scan entire file for signature-like lines
        for _, line := range strings.Split(string(data), "\n") {
            l := strings.TrimSpace(line)
            if strings.HasPrefix(l, "package ") || strings.HasPrefix(l, "import ") {
                continue
            }
            if sigLine.MatchString(l) {
                add(line)
            }
        }
        out[pkg] = items
    }
    return out, nil
}

func startsWithExportedName(line string) bool {
    line = strings.TrimSpace(line)
    if line == "" { return false }
    r := rune(line[0])
    return r >= 'A' && r <= 'Z' && strings.Contains(line, "(")
}

func simplifyDocSig(line string) string {
    s := strings.TrimSpace(line)
    // remove "func " prefix
    s = strings.TrimPrefix(s, "func ")
    // Capture receiver if any: (T) or (*T)
    recv := ""
    if strings.HasPrefix(s, "(") {
        // receiver present
        idx := strings.Index(s, ")")
        if idx > 0 {
            inside := s[1:idx]
            parts := strings.Fields(inside)
            typ := parts[len(parts)-1]
            typ = strings.TrimPrefix(typ, "*")
            recv = typ
            s = strings.TrimSpace(s[idx+1:])
        }
    }
    // Now s starts with Name(…)
    // Extract name and parenthesized params
    nameEnd := strings.Index(s, "(")
    if nameEnd <= 0 { return "" }
    name := strings.TrimSpace(s[:nameEnd])
    // Strip generics in function name e.g. NewHashMap[K, V]
    if i := strings.Index(name, "["); i > 0 {
        name = name[:i]
    }
    rest := s[nameEnd:]
    // Find matching closing paren for params
    paramsEnd := findMatching(rest, 0)
    if paramsEnd < 0 { return "" }
    paramsBody := rest[1:paramsEnd]
    // Normalize params to only types
    normParams := normalizeParamList(paramsBody)
    after := strings.TrimSpace(rest[paramsEnd+1:])
    // Optional results normalized to only types
    results := normalizeResults(after)
    sig := "(" + normParams + ")" + results
    if recv != "" {
        return fmt.Sprintf("(%s).%s%s", recv, name, sig)
    }
    return fmt.Sprintf("%s%s", name, sig)
}

func findMatching(s string, open int) int {
    // s[open] should be '('; return index of matching ')'
    depth := 0
    for i := open; i < len(s); i++ {
        switch s[i] {
        case '(':
            depth++
        case ')':
            depth--
            if depth == 0 { return i }
        }
    }
    return -1
}

// keepExportedOnly keeps only docs whose function/method name is exported and whose receiver type (if any) is exported.
func keepExportedOnly(short string) bool {
    // method form: (Recv).Name(...)
    if strings.HasPrefix(short, "(") {
        // find ")"
        idx := strings.Index(short, ")")
        if idx <= 1 || idx+2 >= len(short) {
            return false
        }
        recv := short[1:idx]
        // exported receiver?
        if recv == "" || !(recv[0] >= 'A' && recv[0] <= 'Z') {
            return false
        }
        rest := short[idx+2:]
        // Name starts at rest until next '('
        nameEnd := strings.Index(rest, "(")
        if nameEnd <= 0 { return false }
        name := rest[:nameEnd]
        return name != "" && (name[0] >= 'A' && name[0] <= 'Z')
    }
    // function form: Name(...)
    nameEnd := strings.Index(short, "(")
    if nameEnd <= 0 { return false }
    name := short[:nameEnd]
    return name != "" && (name[0] >= 'A' && name[0] <= 'Z')
}

// normalizeResults converts a results segment (possibly empty) to canonical types-only string.
func normalizeResults(after string) string {
    if strings.TrimSpace(after) == "" {
        return ""
    }
    out := strings.TrimSpace(after)
    if strings.HasPrefix(out, "(") {
        end := findMatching(out, 0)
        if end < 0 { return "" }
        body := out[1:end]
        parts := splitTopLevel(body)
        types := make([]string, 0, len(parts))
        for _, p := range parts {
            t, count := extractTypeAndCount(p)
            for i := 0; i < count; i++ {
                if t != "" { types = append(types, t) }
            }
        }
        return " (" + strings.Join(types, ", ") + ")"
    }
    return " " + normalizeParamOrType(out)
}

// normalizeParamList transforms a comma-separated parameter list into types-only representation.
func normalizeParamList(body string) string {
    parts := splitTopLevel(body)
    out := make([]string, 0, len(parts))
    for _, p := range parts {
        p = strings.TrimSpace(p)
        if p == "" { continue }
        // Expand grouped params like: a, b string -> string, string
        t, count := extractTypeAndCount(p)
        for i := 0; i < count; i++ {
            out = append(out, t)
        }
    }
    return strings.Join(out, ", ")
}

// splitTopLevel splits by commas ignoring nested parentheses and brackets.
func splitTopLevel(s string) []string {
    var parts []string
    depthParen, depthBrack := 0, 0
    start := 0
    for i := 0; i < len(s); i++ {
        switch s[i] {
        case '(':
            depthParen++
        case ')':
            if depthParen > 0 { depthParen-- }
        case '[':
            depthBrack++
        case ']':
            if depthBrack > 0 { depthBrack-- }
        case ',':
            if depthParen == 0 && depthBrack == 0 {
                parts = append(parts, s[start:i])
                start = i+1
            }
        }
    }
    parts = append(parts, s[start:])
    return parts
}

// extractTypeAndCount returns the type string and how many times it should repeat.
// Handles patterns like "a, b string" (returns "string", 2), "size ...int" ("...int",1), or just "[]int" ("[]int",1)
func extractTypeAndCount(p string) (string, int) {
    p = strings.TrimSpace(p)
    // function type
    if strings.HasPrefix(p, "func") {
        return normalizeFuncType(p), 1
    }
    // If contains varargs
    if strings.Contains(p, "...") {
        // type follows ...
        idx := strings.Index(p, "...")
        t := strings.TrimSpace(p[idx:])
        return removeGenerics(t), 1
    }
    // Find last space outside nested syntax to separate names from type
    lastSpace := lastSpaceTopLevel(p)
    if lastSpace == -1 {
        // No names, return as-is type (after removing generics in identifiers)
        return removeGenerics(strings.TrimSpace(p)), 1
    }
    names := strings.TrimSpace(p[:lastSpace])
    typ := strings.TrimSpace(p[lastSpace+1:])
    // Count names by comma
    c := 1
    if names != "" {
        c = 1 + strings.Count(names, ",")
    }
    return removeGenerics(typ), c
}

func lastSpaceTopLevel(s string) int {
    depthParen, depthBrack := 0, 0
    for i := len(s)-1; i >= 0; i-- {
        switch s[i] {
        case ')': depthParen++
        case '(': if depthParen > 0 { depthParen-- }
        case ']': depthBrack++
        case '[': if depthBrack > 0 { depthBrack-- }
        case ' ':
            if depthParen == 0 && depthBrack == 0 {
                return i
            }
        }
    }
    return -1
}

func normalizeParamOrType(s string) string {
    s = strings.TrimSpace(s)
    if s == "" { return "" }
    if strings.HasPrefix(s, "func") {
        return normalizeFuncType(s)
    }
    return removeGenerics(s)
}

func normalizeFuncType(s string) string {
    // expect s to start with "func"
    t := strings.TrimSpace(strings.TrimPrefix(s, "func"))
    if !strings.HasPrefix(t, "(") { return "func()" }
    end := findMatching(t, 0)
    if end < 0 { return "func()" }
    paramsBody := t[1:end]
    params := normalizeParamList(paramsBody)
    after := strings.TrimSpace(t[end+1:])
    res := normalizeResults(after)
    return "func(" + params + ")" + res
}

// removeGenerics removes generic brackets for exported type identifiers only, e.g., Maper[K,V] -> Maper
// It preserves lowercase tokens like map[K]V and slice types.
func removeGenerics(s string) string {
    // Also normalize inner func types
    if strings.HasPrefix(s, "func") {
        return normalizeFuncType(s)
    }
    // Strip receiver-like identifier names in types "c *Context" -> "*Context"
    if i := lastSpaceTopLevel(s); i != -1 {
        // If left appears to be an identifier (lowercase or uppercase), keep only the right side type
        left := strings.TrimSpace(s[:i])
        right := strings.TrimSpace(s[i+1:])
        if left != "" && right != "" {
            s = right
        }
    }
    // Remove generics for exported identifiers
    re := regexp.MustCompile(`\b([A-Z][A-Za-z0-9_]*)\[[^\]]*\]`)
    for {
        ns := re.ReplaceAllString(s, "$1")
        if ns == s { break }
        s = ns
    }
    return s
}
