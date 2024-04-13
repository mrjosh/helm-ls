package helmdocs

import "github.com/mrjosh/helm-ls/internal/util"

type HelmDocumentation struct {
	Name   string
	Detail string
	Doc    string
}

var (
	BuiltInObjects = []HelmDocumentation{
		{"Values", ".Values", `The values made available through values.yaml, --set and -f.`},
		{"Chart", ".Chart", "Chart metadata"},
		{"Subcharts", ".Subcharts", "This provides access to the scope (.Values, .Charts, .Releases etc.) of subcharts to the parent. For example .Subcharts.mySubChart.myValue to access the myValue in the mySubChart chart."},
		{"Files", ".Files.Get $str", "access non-template files within the chart"},
		{"Capabilities", ".Capabilities.KubeVersion ", "access capabilities of Kubernetes"},
		{"Release", ".Release", `Built-in release values. Attributes include:
    - .Release.Name: Name of the release
    - .Release.Time: Time release was executed
    - .Release.Namespace: Namespace into which release will be placed (if not overridden)
    - .Release.Service: The service that produced this release. Usually Tiller.
    - .Release.IsUpgrade: True if this is an upgrade
    - .Release.IsInstall: True if this is an install
    - .Release.Revision: The revision number
    `},
		{"Template", ".Template", "Contains information about the current template that is being executed"},
	}
	BuiltinFuncs = []HelmDocumentation{
		{"template", "template $str $ctx", "render the template at location $str"},
		{"define", "define $str", "define a template with the name $str"},
		{"and", "and $a $b ...", "if $a then $b else $a"},
		{"call", "call $func $arg $arg2 ...", "call a $func with all $arg(s)"},
		{"html", "html $str", "escape $str for injection into HTML"},
		{"index", "index $collection $key $key2 ...", "get item out of (nested) collection"},
		{"js", "js $str", "encode $str for embedding in JavaScript"},
		{"len", "len $countable", "get the length of a $countable object (list, string, etc)"},
		{"not", "not $x", "negate the boolean value of $x"},
		{"or", "or $a $b", "if $a then $a else $b"},
		{"print", "print $val", "print value"},
		{"printf", "printf $format $val ...", "print $format, injecting values. Follows Sprintf conventions."},
		{"println", "println $val", "print $val followed by newline"},
		{"urlquery", "urlquery $val", "escape value for injecting into a URL query string"},
		{"ne", "ne $a $b", "returns true if $a != $b"},
		{"eq", "eq $a $b ...", "returns true if $a == $b (== ...)"},
		{"lt", "lt $a $b", "returns true if $a < $b"},
		{"gt", "gt $a $b", "returns true if $a > $b"},
		{"le", "le $a $b", "returns true if $a <= $b"},
		{"ge", "ge $a $b", "returns true if $a >= $b"},
	}
	SprigFuncs = []HelmDocumentation{
		// 2.12.0
		{"snakecase", "snakecase $str", "Convert $str to snake_case"},
		{"camelcase", "camelcase $str", "convert string to camelCase"},
		{"shuffle", "shuffle $str", "randomize a string"},
		{"fail", `fail $msg`, "cause the template render to fail with a message $msg."},

		// String
		{"trim", "trim $str", "remove space from either side of string"},
		{"trimAll", "trimAll $trim $str", "remove $trim from either side of $str"},
		{"trimSuffix", "trimSuffix $suf $str", "trim suffix from string"},
		{"trimPrefix", "trimPrefix $pre $str", "trim prefix from string"},
		{"upper", "upper $str", "convert string to uppercase"},
		{"lower", "lower $str", "convert string to lowercase"},
		{"title", "title $str", "convert string to title case"},
		{"untitle", "untitle $str", "convert string from title case"},
		{"substr", "substr $start $len $string", "get a substring of $string, starting at $start and reading $len characters"},
		{"repeat", "repeat $count $str", "repeat $str $count times"},
		{"nospace", "nospace $str", "remove space from inside a string"},
		{"trunc", "trunc $max $str", "truncate $str at $max characters"},
		{"abbrev", "abbrev $max $str", "truncate $str with elipses at max length $max"},
		{"abbrevboth", "abbrevboth $left $right $str", "abbreviate both $left and $right sides of $string"},
		{"initials", "initials $str", "create a string of first characters of each word in $str"},
		{"randAscii", "randAscii", "generate a random string of ASCII characters"},
		{"randNumeric", "randNumeric", "generate a random string of numeric characters"},
		{"randAlpha", "randAlpha", "generate a random string of alphabetic ASCII characters"},
		{"randAlphaNum", "randAlphaNum", "generate a random string of ASCII alphabetic and numeric characters"},
		{"wrap", "wrap $col $str", "wrap $str text at $col columns"},
		{"wrapWith", "wrapWith $col $wrap $str", "wrap $str with $wrap ending each line at $col columns"},
		{"contains", "contains $needle $haystack", "returns true if string $needle is present in $haystack"},
		{"hasPrefix", "hasPrefix $pre $str", "returns true if $str begins with $pre"},
		{"hasSuffix", "hasSuffix $suf $str", "returns true if $str ends with $suf"},
		{"quote", "quote $str", "surround $str with double quotes (\")"},
		{"squote", "squote $str", "surround $str with single quotes (')"},
		{"cat", "cat $str1 $str2 ...", "concatenate all given strings into one, separated by spaces"},
		{"indent", "indent $count $str", "indent $str with $count space chars on the left"},
		{"nindent", "nindent $count $str", "indent $str with $count space chars on the left and prepend a new line to $str"},
		{"replace", "replace $find $replace $str", "find $find and replace with $replace"},

		// String list
		{"plural", "plural $singular $plural $count", "if $count is 1, return $singular, else return $plural"},
		{"join", "join $sep $list", "concatenate list of strings into one, separated by $sep"},
		{"splitList", "splitList $sep $str", "split $str into a list of strings, separating at $sep"},
		{"split", "split $sep $str", "split $str on $sep and store results in a dictionary"},
		{"sortAlpha", "sortAlpha $strings", "sort a list of strings into alphabetical order"},
		// Math
		{"add", "add $a $b $c", "add two or more numbers"},
		{"add1", "add1 $a", "increment $a by 1"},
		{"sub", "sub $a $b", "subtract $a from $b"},
		{"div", "div $a $b", "divide $b by $a"},
		{"mod", "mod $a $b", "modulo $b by $a"},
		{"mul", "mult $a $b", "multiply $b by $a"},
		{"max", "max $a $b ...", "return max integer"},
		{"min", "min $a $b ...", "return min integer"},
		// Integer list
		{"until", "until $count", "return a list of integers beginning with 0 and ending with $until - 1"},
		{"untilStep", "untilStep $start $max $step", "start with $start, and add $step until reaching $max"},
		// Date
		{"now", "now", "current date/time"},
		{"date", "date $format $date", "Format a $date with format string $format"},
		{"dateInZone", "date $format $date $tz", "Format $date with $format in timezone $tz"},
		{"dateModify", "dateModify $mod $date", "Modify $day by string $mod"},
		{"htmlDate", "htmlDate $date", "format $date accodring to HTML5 date format"},
		{"htmlDateInZone", "$htmlDate $date $tz", "format $date in $tz for HTML5 date fields"},
		// Defaults
		{"default", "default $default $optional", "if $optional is not set, use $default"},
		{"empty", "empty $val", "if $value is empty, return true. Otherwise return false"},
		{"coalesce", "coalesce $val1 $val2 ...", "for a list of values, return the first non-empty one"},
		{"ternary", "ternary $then $else $condition", "if $condition is true, return $then. Otherwise return $else"},
		// Encoding
		{"b64enc", "b64enc $str", "encode $str with base64 encoding"},
		{"b64dec", "b64dec $str", "decode $str with base64 decoder"},
		{"b32enc", "b32enc $str", "encode $str with base32 encoder"},
		{"b32dec", "b32dec $str", "decode $str with base32 decoder"},
		// Lists
		{"list", "list $a $b ...", "create a list from all args"},
		{"first", "first $list", "return the first item in a $list"},
		{"rest", "rest $list", "return all but the first of $list"},
		{"last", "last $list", "return last item in $list"},
		{"initial", "initial $list", "return all but last in $list"},
		{"append", "append $list $item", "append $item to $list"},
		{"prepend", "prepend $list $item", "prepend $item to $list"},
		{"reverse", "reverse $list", "reverse $list item order"},
		{"uniq", "uniq $list", "remove duplicates from list"},
		{"without", "without $list $item ...", "return $list with $item(s) removed"},
		{"has", "has $item $list", "return true if $item is in $list"},
		// Dictionaries
		{"dict", "dict $key $val $key2 $val2 ...", "create dictionary with $key/$val pairs"},
		{"set", "set $dict $key $val", "set $key=$val in $dict (mutates dict)"},
		{"unset", "unset $dict $key", "remove $key from $dict"},
		{"hasKey", "hasKey $dict $key", "returns true if $key is in $dict"},
		{"pluck", "pluck $key $dict1 $dict2 ...", "Get same $key from all $dict(s)"},
		{"merge", "merge $dest $src", "deeply merge $src into $dest"},
		{"keys", "keys $dict", "get list of all keys in dict. Keys are not ordered."},
		{"pick", "pick $dict $key1 $key2 ...", "extract $key(s) from $dict and create new dict with just those key/val pairs"},
		{"omit", "omit $dict $key1 $key2...", "return new dict with $key(s) removed from $dict"},
		// Type Conversion
		{"atoi", "atoi $str", "convert $str to integer. Zero if conversion fails."},
		{"float64", "float64 $val", "convert $val to float64"},
		{"int", "int $val", "convert $val to int"},
		{"int64", "int64 $val", "convert $val to int64"},
		{"toString", "toString $val", "convert $val to string"},
		{"toStrings", "toStrings $list", "convert every item in $list to string, return list of strings"},
		// File Path
		{"base", "base $path", "return base name (last element) of $path"},
		{"dir", "dir $path", "return all but base name of path (return next dir up)"},
		{"clean", "clean $path", "clean up the $path"},
		{"ext", "ext $path", "return the file extension (or empty string) of last item on $path"},
		{"isAbs", "isAps $path", "return true if $path is absolute"},
		// UUID
		{"uuidv4", "uuidv4", "generate a UUID v4 (random universally unique ID"},
		// OS
		{"env", "env $var", "(UNSUPPORTED IN HELM) get env var"},
		{"expandenv", "expandenv $str", "(UNSUPPORTED IN HELM) expand env vars in string"},
		// SemVer
		{"semver", "semver $version", "parse a SemVer string (1.2.3-alpha.4+1234). [Reference](http://masterminds.github.io/sprig/semver.html)"},
		{"semverCompare", "semverCompare $ver1 $ver2", "Compare $ver1 and $ver2. $ver1 can be a [SemVer range]((http://masterminds.github.io/sprig/semver.html)."},
		// Reflection
		{"kindOf", "kindOf $val", "return the Go kind (primitive type) of a value"},
		{"kindIs", "kindIs $kind $val", "returns true if $val is of kind $kind"},
		{"typeOf", "typeOf $val", "returns a string indicate the type of $val"},
		{"typeIs", "typeIs $type $val", "returns true if $val is of type $type"},
		{"typeIsLike", "typeIsLike $substr $val", "returns true if $substr is found in $val's type"},
		// Crypto
		{"sha1sum", "sha1sum $str", "generate a SHA-1 sum of $str"},
		{"sha256sum", "sha256sum $str", "generate a SHA-256 sum of $str"},
		{"derivePassword", "derivePassword $counter $long $pass $user $domain", "generate a password from [Master Password](http://masterpasswordapp.com/algorithm.html) spec"},
		{"generatePrivateKey", "generatePrivateKey 'ecdsa'", "generate private PEM key (takes dsa, rsa, or ecdsa)"},
	}
	HelmFuncs = []HelmDocumentation{
		{"include", "include $str $ctx", "(chainable) include the named template with the given context."},
		{"toYaml", "toYaml $var", "convert $var to YAML"},
		{"toJson", "toJson $var", "convert $var to JSON"},
		{"toToml", "toToml $var", "convert $var to TOML"},
		{"fromYaml", "fromYaml $str", "parse YAML into a dict or list"},
		{"fromJson", "fromJson $str", "parse JSON $str into a dict or list"},
		{"required", "required $str $val", "fail template with message $str if $val is not provided or is empty"},
	}

	CapabilitiesVals = []HelmDocumentation{
		{"TillerVersion", ".Capabilities.TillerVersion", "Tiller version"},
		{"APIVersions", "Capabilities.APIVersions", "A set of versions."},
		{"APIVersions.Has", "Capabilities.APIVersions.Has $version", "Indicates whether a version (e.g., batch/v1) or resource (e.g., apps/v1/Deployment) is available on the cluster."},
		{"KubeVersion", "Capabilities.KubeVersion", "The Kubernetes version."},
		{"KubeVersion.Version", "Capabilities.KubeVersion.Version", "The Kubernetes version in semver format."},
		{"KubeVersion.Major", "Capabilities.KubeVersion.Major", "The Kubernetes major version."},
		{"KubeVersion.Minor", "Capabilities.KubeVersion.Minor", "The Kubernetes minor version."},
		{"KubeVersion.GitCommit", "Capabilities.HelmVersion", "The object containing the Helm Version details, it is the same output of helm version."},
		{"KubeVersion.GitTreeState", "Capabilities.HelmVersion.Version", "The current Helm version in semver format."},
		{"HelmVersion.GitCommit", "Capabilities.HelmVersion.GitCommit", "The Helm git sha1."},
		{"HelmVersion.GitTreeState", "Capabilities.HelmVersion.GitTreeState", "The state of the Helm git tree."},
		{"HelmVersion.GoVersion", "Capabilities.HelmVersion.GoVersion", "The version of the Go compiler used."},
	}

	ChartVals = []HelmDocumentation{
		{"Name", ".Chart.Name", "Name of the chart"},
		{"Version", ".Chart.Version", "Version of the chart"},
		{"Description", ".Chart.Description", "Chart description"},
		{"Keywords", ".Chart.Keywords", "A list of keywords (as strings)"},
		{"Home", ".Chart.Home", "The chart homepage URL"},
		{"Sources", ".Chart.Sources", "A list of chart download URLs"},
		{"Maintainers", ".Chart.Maintainers", "list of maintainer objects"},
		{"Icon", ".Chart.Icon", "The URL to the chart's icon file"},
		{"AppVersion", ".Chart.AppVersion", "The version of the main app contained in this chart"},
		{"Deprecated", ".Chart.Deprecated", "If true, this chart is no longer maintained"},
		{"TillerVersion", ".Chart.TillerVersion", "The version (range) if Tiller that this chart can run on."},
		{"APIVersion", ".Chart.APIVersion", "The API Version of this chart"},
		{"Condition", ".Chart.Condition", "The condition to check to enable chart"},
		{"Tags", ".Chart.Tags", "The tags to check to enable chart"},
		{"Annotations", ".Chart.Annotations", "Additional annotations (key-value pairs)"},
		{"KubeVersion", ".Chart.KubeVersion", "Kubernetes version required"},
		{"Dependencies", ".Chart.Dependencies", "List of chart dependencies"},
		{"Type", ".Chart.Type", "Chart type (application or library)"},
	}

	TemplateVals = []HelmDocumentation{
		{"Name", ".Template.Name", "A namespaced file path to the current template (e.g. mychart/templates/mytemplate.yaml)"},
		{"BasePath", ".Template.BasePath", "The namespaced path to the templates directory of the current chart (e.g. mychart/templates)."},
	}

	ReleaseVals = []HelmDocumentation{
		{"Name", ".Release.Name", "Name of the release"},
		{"Time", ".Release.Time", "Time of the release"},
		{"Namespace", ".Release.Namespace", "Default namespace of the release"},
		{"Service", ".Release.Service", "The service that is rendering the present template. On Helm, this is always Helm"},
		{"IsUpgrade", ".Release.IsUpgrade", "True if this is an upgrade operation"},
		{"IsInstall", ".Release.IsInstall", "True if this is an install operation"},
		{"Revision", ".Release.Revision", "Release revision number (starts at 1)"},
	}

	FilesVals = []HelmDocumentation{
		{"Get", ".Files.Get $path", "Get file contents. Path is relative to chart."},
		{"GetBytes", ".Files.GetBytes $path", "Get file contents as a byte array. Path is relative to chart."},
		{"Glob", ".Files.Glob $glob", "Returns a list of files whose names match the given shell glob pattern."},
		{"Lines", ".Files.Lines $path", "Reads a file line-by-line. This is useful for iterating over each line in a file."},
		{"AsSecrets", ".Files.AsSecrets $path", "Returns the file bodies as Base 64 encoded strings."},
		{"AsConfig", ".Files.AsConfig $path", "Returns file bodies as a YAML map."},
	}

	BuiltInOjectVals = map[string][]HelmDocumentation{
		"Chart":        ChartVals,
		"Release":      ReleaseVals,
		"Files":        FilesVals,
		"Capabilities": CapabilitiesVals,
		"Template":     TemplateVals,
	}
)

func GetFunctionByName(name string) (HelmDocumentation, bool) {
	completionItems := [][]HelmDocumentation{
		BuiltinFuncs,
		SprigFuncs,
		HelmFuncs,
	}
	toSearch := util.ConcatMultipleSlices(completionItems)
	for _, completionItem := range toSearch {
		if name == completionItem.Name {
			return completionItem, true
		}
	}
	return HelmDocumentation{}, false
}
