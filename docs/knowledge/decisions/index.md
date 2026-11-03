# Decisions
Architecture decision records. `decision: proposed` until the maintainer agrees.

* [DEC-001 Mask and reuse go/parser](001-mask-and-reuse-go-parser.md) - parse by masking markup into a same-length Go skeleton and grafting markup nodes that embed go/ast nodes
* [DEC-002 Element is an alias](002-element-is-an-alias.md) - golpe.Element = func(io.Writer) error, so writer funcs interoperate without conversions
* [DEC-003 template is a contextual keyword](003-template-is-a-contextual-keyword.md) - keyword only at top-level declaration position, so all Go stays valid
* [DEC-004 Runtime dispatch for interpolation](004-runtime-dispatch-for-interpolation.md) - {expr} calls a runtime helper that renders Elements and escapes everything else
