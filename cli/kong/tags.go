package kong

// Kong struct tag keys.
const (
	tagAliases     = "aliases"
	tagArg         = "arg"
	tagCmd         = "cmd"
	tagClib        = "clib"
	tagDefault     = "default"
	tagEnum        = "enum"
	tagHelp        = "help"
	tagHidden      = "hidden"
	tagName        = "name"
	tagNegatable   = "negatable"
	tagOptional    = "optional"
	tagPlaceholder = "placeholder"
	tagPredictor   = "predictor"
	tagShort       = "short"
	tagShowAliases = "show-aliases"
	tagType        = "type"
)

// Kong struct tag values.
const (
	kongTypeCounter      = "counter"
	kongTypeExistingDir  = "existingdir"
	kongTypeExistingFile = "existingfile"
	kongTypeFileContent  = "filecontent"
	kongTypePath         = "path"
	predictorPath        = "path"
)
