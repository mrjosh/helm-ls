package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"strings"

	"github.com/mrjosh/helm-ls/pkg/chartutil"
	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"
	yaml "gopkg.in/yaml.v2"
)

var (
	emptyItems = make([]lsp.CompletionItem, 0)
	basicItems = []lsp.CompletionItem{
		variableCompletionItem("Values", ".Values", `The values made available through values.yaml, --set and -f.`),
		variableCompletionItem("Chart", ".Chart", "Chart metadata"),
		variableCompletionItem("Files", ".Files.Get $str", "access non-template files within the chart"),
		variableCompletionItem("Capabilities", ".Capabilities.KubeVersion ", "access capabilities of Kubernetes"),
		variableCompletionItem("Release", ".Release", `Built-in release values. Attributes include:
    - .Release.Name: Name of the release
    - .Release.Time: Time release was executed
    - .Release.Namespace: Namespace into which release will be placed (if not overridden)
    - .Release.Service: The service that produced this release. Usually Tiller.
    - .Release.IsUpgrade: True if this is an upgrade
    - .Release.IsInstall: True if this is an install
    - .Release.Revision: The revision number
    `),
	}
	builtinFuncs = []lsp.CompletionItem{
		functionCompletionItem("template", "template $str $ctx", "render the template at location $str"),
		functionCompletionItem("define", "define $str", "define a template with the name $str"),
		functionCompletionItem("and", "and $a $b ...", "if $a then $b else $a"),
		functionCompletionItem("call", "call $func $arg $arg2 ...", "call a $func with all $arg(s)"),
		functionCompletionItem("html", "html $str", "escape $str for injection into HTML"),
		functionCompletionItem("index", "index $collection $key $key2 ...", "get item out of (nested) collection"),
		functionCompletionItem("js", "js $str", "encode $str for embedding in JavaScript"),
		functionCompletionItem("len", "len $countable", "get the length of a $countable object (list, string, etc)"),
		functionCompletionItem("not", "not $x", "negate the boolean value of $x"),
		functionCompletionItem("or", "or $a $b", "if $a then $a else $b"),
		functionCompletionItem("print", "print $val", "print value"),
		functionCompletionItem("printf", "printf $format $val ...", "print $format, injecting values. Follows Sprintf conventions."),
		functionCompletionItem("println", "println $val", "print $val followed by newline"),
		functionCompletionItem("urlquery", "urlquery $val", "escape value for injecting into a URL query string"),
		functionCompletionItem("ne", "ne $a $b", "returns true if $a != $b"),
		functionCompletionItem("eq", "eq $a $b ...", "returns true if $a == $b (== ...)"),
		functionCompletionItem("lt", "lt $a $b", "returns true if $a < $b"),
		functionCompletionItem("gt", "gt $a $b", "returns true if $a > $b"),
		functionCompletionItem("le", "le $a $b", "returns true if $a <= $b"),
		functionCompletionItem("ge", "ge $a $b", "returns true if $a >= $b"),
	}
	sprigFuncs = []lsp.CompletionItem{
		// 2.12.0
		functionCompletionItem("snakecase", "snakecase $str", "Convert $str to snake_case"),
		functionCompletionItem("camelcase", "camelcase $str", "convert string to camelCase"),
		functionCompletionItem("shuffle", "shuffle $str", "randomize a string"),
		functionCompletionItem("fail", `fail $msg`, "cause the template render to fail with a message $msg."),

		// String
		functionCompletionItem("trim", "trim $str", "remove space from either side of string"),
		functionCompletionItem("trimAll", "trimAll $trim $str", "remove $trim from either side of $str"),
		functionCompletionItem("trimSuffix", "trimSuffix $suf $str", "trim suffix from string"),
		functionCompletionItem("trimPrefix", "trimPrefix $pre $str", "trim prefix from string"),
		functionCompletionItem("upper", "upper $str", "convert string to uppercase"),
		functionCompletionItem("lower", "lower $str", "convert string to lowercase"),
		functionCompletionItem("title", "title $str", "convert string to title case"),
		functionCompletionItem("untitle", "untitle $str", "convert string from title case"),
		functionCompletionItem("substr", "substr $start $len $string", "get a substring of $string, starting at $start and reading $len characters"),
		functionCompletionItem("repeat", "repeat $count $str", "repeat $str $count times"),
		functionCompletionItem("nospace", "nospace $str", "remove space from inside a string"),
		functionCompletionItem("trunc", "trunc $max $str", "truncate $str at $max characters"),
		functionCompletionItem("abbrev", "abbrev $max $str", "truncate $str with elipses at max length $max"),
		functionCompletionItem("abbrevboth", "abbrevboth $left $right $str", "abbreviate both $left and $right sides of $string"),
		functionCompletionItem("initials", "initials $str", "create a string of first characters of each word in $str"),
		functionCompletionItem("randAscii", "randAscii", "generate a random string of ASCII characters"),
		functionCompletionItem("randNumeric", "randNumeric", "generate a random string of numeric characters"),
		functionCompletionItem("randAlpha", "randAlpha", "generate a random string of alphabetic ASCII characters"),
		functionCompletionItem("randAlphaNum", "randAlphaNum", "generate a random string of ASCII alphabetic and numeric characters"),
		functionCompletionItem("wrap", "wrap $col $str", "wrap $str text at $col columns"),
		functionCompletionItem("wrapWith", "wrapWith $col $wrap $str", "wrap $str with $wrap ending each line at $col columns"),
		functionCompletionItem("contains", "contains $needle $haystack", "returns true if string $needle is present in $haystack"),
		functionCompletionItem("hasPrefix", "hasPrefix $pre $str", "returns true if $str begins with $pre"),
		functionCompletionItem("hasSuffix", "hasSuffix $suf $str", "returns true if $str ends with $suf"),
		functionCompletionItem("quote", "quote $str", "surround $str with double quotes (\")"),
		functionCompletionItem("squote", "squote $str", "surround $str with single quotes (')"),
		functionCompletionItem("cat", "cat $str1 $str2 ...", "concatenate all given strings into one, separated by spaces"),
		functionCompletionItem("indent", "indent $count $str", "indent $str with $count space chars on the left"),
		functionCompletionItem("nindent", "nindent $count $str", "indent $str with $count space chars on the left and prepend a new line to $str"),
		functionCompletionItem("replace", "replace $find $replace $str", "find $find and replace with $replace"),

		// String list
		functionCompletionItem("plural", "plural $singular $plural $count", "if $count is 1, return $singular, else return $plural"),
		functionCompletionItem("join", "join $sep $list", "concatenate list of strings into one, separated by $sep"),
		functionCompletionItem("splitList", "splitList $sep $str", "split $str into a list of strings, separating at $sep"),
		functionCompletionItem("split", "split $sep $str", "split $str on $sep and store results in a dictionary"),
		functionCompletionItem("sortAlpha", "sortAlpha $strings", "sort a list of strings into alphabetical order"),
		// Math
		functionCompletionItem("add", "add $a $b $c", "add two or more numbers"),
		functionCompletionItem("add1", "add1 $a", "increment $a by 1"),
		functionCompletionItem("sub", "sub $a $b", "subtract $a from $b"),
		functionCompletionItem("div", "div $a $b", "divide $b by $a"),
		functionCompletionItem("mod", "mod $a $b", "modulo $b by $a"),
		functionCompletionItem("mul", "mult $a $b", "multiply $b by $a"),
		functionCompletionItem("max", "max $a $b ...", "return max integer"),
		functionCompletionItem("min", "min $a $b ...", "return min integer"),
		// Integer list
		functionCompletionItem("until", "until $count", "return a list of integers beginning with 0 and ending with $until - 1"),
		functionCompletionItem("untilStep", "untilStep $start $max $step", "start with $start, and add $step until reaching $max"),
		// Date
		functionCompletionItem("now", "now", "current date/time"),
		functionCompletionItem("date", "date $format $date", "Format a $date with format string $format"),
		functionCompletionItem("dateInZone", "date $format $date $tz", "Format $date with $format in timezone $tz"),
		functionCompletionItem("dateModify", "dateModify $mod $date", "Modify $day by string $mod"),
		functionCompletionItem("htmlDate", "htmlDate $date", "format $date accodring to HTML5 date format"),
		functionCompletionItem("htmlDateInZone", "$htmlDate $date $tz", "format $date in $tz for HTML5 date fields"),
		// Defaults
		functionCompletionItem("default", "default $default $optional", "if $optional is not set, use $default"),
		functionCompletionItem("empty", "empty $val", "if $value is empty, return true. Otherwise return false"),
		functionCompletionItem("coalesce", "coalesce $val1 $val2 ...", "for a list of values, return the first non-empty one"),
		functionCompletionItem("ternary", "ternary $then $else $condition", "if $condition is true, return $then. Otherwise return $else"),
		// Encoding
		functionCompletionItem("b64enc", "b64enc $str", "encode $str with base64 encoding"),
		functionCompletionItem("b64dec", "b64dec $str", "decode $str with base64 decoder"),
		functionCompletionItem("b32enc", "b32enc $str", "encode $str with base32 encoder"),
		functionCompletionItem("b32dec", "b32dec $str", "decode $str with base32 decoder"),
		// Lists
		functionCompletionItem("list", "list $a $b ...", "create a list from all args"),
		functionCompletionItem("first", "first $list", "return the first item in a $list"),
		functionCompletionItem("rest", "rest $list", "return all but the first of $list"),
		functionCompletionItem("last", "last $list", "return last item in $list"),
		functionCompletionItem("initial", "initial $list", "return all but last in $list"),
		functionCompletionItem("append", "append $list $item", "append $item to $list"),
		functionCompletionItem("prepend", "prepend $list $item", "prepend $item to $list"),
		functionCompletionItem("reverse", "reverse $list", "reverse $list item order"),
		functionCompletionItem("uniq", "uniq $list", "remove duplicates from list"),
		functionCompletionItem("without", "without $list $item ...", "return $list with $item(s) removed"),
		functionCompletionItem("has", "has $item $list", "return true if $item is in $list"),
		// Dictionaries
		functionCompletionItem("dict", "dict $key $val $key2 $val2 ...", "create dictionary with $key/$val pairs"),
		functionCompletionItem("set", "set $dict $key $val", "set $key=$val in $dict (mutates dict)"),
		functionCompletionItem("unset", "unset $dict $key", "remove $key from $dict"),
		functionCompletionItem("hasKey", "hasKey $dict $key", "returns true if $key is in $dict"),
		functionCompletionItem("pluck", "pluck $key $dict1 $dict2 ...", "Get same $key from all $dict(s)"),
		functionCompletionItem("merge", "merge $dest $src", "deeply merge $src into $dest"),
		functionCompletionItem("keys", "keys $dict", "get list of all keys in dict. Keys are not ordered."),
		functionCompletionItem("pick", "pick $dict $key1 $key2 ...", "extract $key(s) from $dict and create new dict with just those key/val pairs"),
		functionCompletionItem("omit", "omit $dict $key1 $key2...", "return new dict with $key(s) removed from $dict"),
		// Type Conversion
		functionCompletionItem("atoi", "atoi $str", "convert $str to integer. Zero if conversion fails."),
		functionCompletionItem("float64", "float64 $val", "convert $val to float64"),
		functionCompletionItem("int", "int $val", "convert $val to int"),
		functionCompletionItem("int64", "int64 $val", "convert $val to int64"),
		functionCompletionItem("toString", "toString $val", "convert $val to string"),
		functionCompletionItem("toStrings", "toStrings $list", "convert every item in $list to string, return list of strings"),
		// File Path
		functionCompletionItem("base", "base $path", "return base name (last element) of $path"),
		functionCompletionItem("dir", "dir $path", "return all but base name of path (return next dir up)"),
		functionCompletionItem("clean", "clean $path", "clean up the $path"),
		functionCompletionItem("ext", "ext $path", "return the file extension (or empty string) of last item on $path"),
		functionCompletionItem("isAbs", "isAps $path", "return true if $path is absolute"),
		// UUID
		functionCompletionItem("uuidv4", "uuidv4", "generate a UUID v4 (random universally unique ID"),
		// OS
		functionCompletionItem("env", "env $var", "(UNSUPPORTED IN HELM) get env var"),
		functionCompletionItem("expandenv", "expandenv $str", "(UNSUPPORTED IN HELM) expand env vars in string"),
		// SemVer
		functionCompletionItem("semver", "semver $version", "parse a SemVer string (1.2.3-alpha.4+1234). [Reference](http://masterminds.github.io/sprig/semver.html)"),
		functionCompletionItem("semverCompare", "semverCompare $ver1 $ver2", "Compare $ver1 and $ver2. $ver1 can be a [SemVer range]((http://masterminds.github.io/sprig/semver.html)."),
		// Reflection
		functionCompletionItem("kindOf", "kindOf $val", "return the Go kind (primitive type) of a value"),
		functionCompletionItem("kindIs", "kindIs $kind $val", "returns true if $val is of kind $kind"),
		functionCompletionItem("typeOf", "typeOf $val", "returns a string indicate the type of $val"),
		functionCompletionItem("typeIs", "typeIs $type $val", "returns true if $val is of type $type"),
		functionCompletionItem("typeIsLike", "typeIsLike $substr $val", "returns true if $substr is found in $val's type"),
		// Crypto
		functionCompletionItem("sha1sum", "sha1sum $str", "generate a SHA-1 sum of $str"),
		functionCompletionItem("sha256sum", "sha256sum $str", "generate a SHA-256 sum of $str"),
		functionCompletionItem("derivePassword", "derivePassword $counter $long $pass $user $domain", "generate a password from [Master Password](http://masterpasswordapp.com/algorithm.html) spec"),
		functionCompletionItem("generatePrivateKey", "generatePrivateKey 'ecdsa'", "generate private PEM key (takes dsa, rsa, or ecdsa)"),
	}
	helmFuncs = []lsp.CompletionItem{
		functionCompletionItem("include", "include $str $ctx", "(chainable) include the named template with the given context."),
		functionCompletionItem("toYaml", "toYaml $var", "convert $var to YAML"),
		functionCompletionItem("toJson", "toJson $var", "convert $var to JSON"),
		functionCompletionItem("toToml", "toToml $var", "convert $var to TOML"),
		functionCompletionItem("fromYaml", "fromYaml $str", "parse YAML into a dict or list"),
		functionCompletionItem("fromJson", "fromJson $str", "parse JSON $str into a dict or list"),
		functionCompletionItem("required", "required $str $val", "fail template with message $str if $val is not provided or is empty"),
	}
	functionsCompletionItems = make([]lsp.CompletionItem, 0)
)

func init() {
	functionsCompletionItems = append(functionsCompletionItems, helmFuncs...)
	functionsCompletionItems = append(functionsCompletionItems, builtinFuncs...)
	functionsCompletionItems = append(functionsCompletionItems, sprigFuncs...)
}

func (h *langHandler) handleTextDocumentCompletion(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) (err error) {

	if req.Params() == nil {
		return &jsonrpc2.Error{Code: jsonrpc2.InvalidParams}
	}

	var params lsp.CompletionParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return err
	}

	doc, ok := h.documents.Get(params.TextDocument.URI)
	if !ok {
		return errors.New("Could not get document: " + params.TextDocument.URI.Filename())
	}

	var (
		word             = doc.WordAt(params.Position)
		splitted         = strings.Split(word, ".")
		items            = make([]lsp.CompletionItem, 0)
		variableSplitted = []string{}
	)

	for _, s := range splitted {
		if s != "" {
			variableSplitted = append(variableSplitted, s)
		}
	}

	logger.Println(fmt.Sprintf("Word < %s >", word))

	if len(variableSplitted) == 0 {
		return reply(ctx, basicItems, err)
	}

	// $ always points to the root context so we can safely remove it
	// as long the LSP does not know about ranges
	if variableSplitted[0] == "$" && len(variableSplitted) > 1 {
		variableSplitted = variableSplitted[1:]
	}

	switch variableSplitted[0] {
	case "Chart":
		items = h.getChartVals()
	case "Values":
		items = h.getValue(h.values, variableSplitted[1:])
	case "Release":
		items = h.getReleaseVals()
	case "Files":
		items = h.getFilesVals()
	case "Capabilities":
		items = h.getCapabilitiesVals()
	default:
		items = basicItems
		items = append(items, functionsCompletionItems...)
	}

	return reply(ctx, items, err)
}

func (h *langHandler) getCapabilitiesVals() []lsp.CompletionItem {
	return []lsp.CompletionItem{
		variableCompletionItem("KubeVersion", ".Capabilities.KubeVersion", "Kubernetes version"),
		variableCompletionItem("TillerVersion", ".Capabilities.TillerVersion", "Tiller version"),
		functionCompletionItem("ApiVersions.Has", `.Capabilities.ApiVersions.Has "batch/v1"`, "Returns true if the given Kubernetes API/version is present on the cluster"),
	}
}

func (h *langHandler) getChartVals() []lsp.CompletionItem {
	return []lsp.CompletionItem{
		variableCompletionItem("Name", ".Chart.Name", "Name of the chart"),
		variableCompletionItem("Version", ".Chart.Version", "Version of the chart"),
		variableCompletionItem("Description", ".Chart.Description", "Chart description"),
		variableCompletionItem("Keywords", ".Chart.Keywords", "A list of keywords (as strings)"),
		variableCompletionItem("Home", ".Chart.Home", "The chart homepage URL"),
		variableCompletionItem("Sources", ".Chart.Sources", "A list of chart download URLs"),
		variableCompletionItem("Maintainers", ".Chart.Maintainers", "list of maintainer objects"),
		variableCompletionItem("Icon", ".Chart.Icon", "The URL to the chart's icon file"),
		variableCompletionItem("AppVersion", ".Chart.AppVersion", "The version of the main app contained in this chart"),
		variableCompletionItem("Deprecated", ".Chart.Deprecated", "If true, this chart is no longer maintained"),
		variableCompletionItem("TillerVersion", ".Chart.TillerVersion", "The version (range) if Tiller that this chart can run on."),
	}
}

func (h *langHandler) getReleaseVals() []lsp.CompletionItem {
	return []lsp.CompletionItem{
		variableCompletionItem("Name", ".Release.Name", "Name of the release"),
		variableCompletionItem("Time", ".Release.Time", "Time of the release"),
		variableCompletionItem("Namespace", ".Release.Namespace", "Default namespace of the release"),
		variableCompletionItem("ServiceName", ".Release.Service", "Name of the service that produced the release (almost always Tiller)"),
		variableCompletionItem("IsUpgrade", ".Release.IsUpgrade", "True if this is an upgrade operation"),
		variableCompletionItem("IsInstall", ".Release.IsInstall", "True if this is an install operation"),
		variableCompletionItem("Revision", ".Release.Revision", "Release revision number (starts at 1)"),
	}
}

func (h *langHandler) getFilesVals() []lsp.CompletionItem {
	return []lsp.CompletionItem{
		functionCompletionItem("Get", ".Files.Get $path", "Get file contents. Path is relative to chart."),
		functionCompletionItem("GetBytes", ".Files.GetBytes $path", "Get file contents as a byte array. Path is relative to chart."),
	}
}

func (h *langHandler) getValue(values chartutil.Values, splittedVar []string) []lsp.CompletionItem {

	var (
		err       error
		items     = make([]lsp.CompletionItem, 0)
		tableName = strings.Join(splittedVar, ".")
	)

	if len(splittedVar) > 0 {

		values, err = values.Table(tableName)
		if err != nil {
			logger.Println(err)
			if errors.Is(err, chartutil.ErrNoTable{}) {
				return emptyItems
			}
			return emptyItems
		}

	}

	for variable, value := range values {
		items = h.setItem(items, value, variable)
	}

	return items
}

func (h *langHandler) setItem(items []lsp.CompletionItem, value interface{}, variable string) []lsp.CompletionItem {

	var (
		itemKind      = lsp.CompletionItemKindVariable
		valueOf       = reflect.ValueOf(value)
		documentation = valueOf.String()
	)

	logger.Println("ValueKind: ", valueOf)

	switch valueOf.Kind() {
	case reflect.Slice, reflect.Map:
		itemKind = lsp.CompletionItemKindStruct
		documentation = h.toYAML(value)
	case reflect.Bool:
		itemKind = lsp.CompletionItemKindVariable
		documentation = h.getBoolType(value)
	case reflect.Float32, reflect.Float64:
		documentation = fmt.Sprintf("%.2f", valueOf.Float())
		itemKind = lsp.CompletionItemKindVariable
	case reflect.Invalid:
		documentation = "<Unknown>"
	default:
		itemKind = lsp.CompletionItemKindField
	}

	return append(items, lsp.CompletionItem{
		Label:         variable,
		InsertText:    variable,
		Documentation: documentation,
		Detail:        valueOf.Kind().String(),
		Kind:          itemKind,
	})
}

func (h *langHandler) toYAML(value interface{}) string {
	valBytes, _ := yaml.Marshal(value)
	return string(valBytes)
}

func (h *langHandler) getBoolType(value interface{}) string {
	if val, ok := value.(bool); ok && val {
		return "True"
	}
	return "False"
}

func variableCompletionItem(name, detail, doc string) lsp.CompletionItem {
	return lsp.CompletionItem{
		Label:      name,
		InsertText: name,
		Detail:     detail,
		Kind:       lsp.CompletionItemKindVariable,
	}
}

func functionCompletionItem(name, detail, doc string) lsp.CompletionItem {
	return lsp.CompletionItem{
		Label:         name,
		InsertText:    name,
		Detail:        detail,
		Documentation: doc,
		Kind:          lsp.CompletionItemKindFunction,
	}
}
