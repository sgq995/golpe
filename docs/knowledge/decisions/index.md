# Decisions
Architecture decision records. `decision: proposed` until the maintainer agrees.

* [DEC-001 Mask and reuse go/parser](001-mask-and-reuse-go-parser.md) - parse by masking markup into a same-length Go skeleton and grafting markup nodes that embed go/ast nodes (accepted)
* [DEC-002 Element is an alias](002-element-is-an-alias.md) - element means func(io.Writer) error; golpe.Element is an optional alias (accepted)
* [DEC-003 template is a contextual keyword](003-template-is-a-contextual-keyword.md) - keyword only at top-level declaration position, so all Go stays valid (accepted)
* [DEC-004 Runtime dispatch for interpolation](004-runtime-dispatch-for-interpolation.md) - superseded by DEC-005
* [DEC-005 Standard library only output](005-standard-library-only-output.md) - generated code imports only fmt, html and io; golpe's API is optional (accepted)
* [DEC-006 writer() and render calls](006-writer-and-render-calls.md) - {expr} containing writer() is a checked render call; everything else is an escaped value (accepted)
* [DEC-007 Component tags](007-component-tags.md) - non-element tags call components; props from the declaration; no children in v1 (accepted)
* [DEC-008 No implicit context](008-no-implicit-context.md) - Element stays func(io.Writer) error; context is an explicit parameter (accepted)
* [DEC-009 Escape by default](009-escape-by-default.md) - values are escaped; static text is verbatim; trusted HTML goes through writer() (accepted)
