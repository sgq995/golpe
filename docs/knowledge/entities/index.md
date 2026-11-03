# Entities
The things Golpe reads, produces and reasons about.

# Source
* [Golpe source file](golpe-file.md) - a .golpe file: Go that may also contain template declarations
* [Template declaration](template-declaration.md) - a top-level template whose body mixes Go and markup and returns func(io.Writer) error
* [Markup node](markup-node.md) - a node of the HTML tree inside a template
* [Component tag](component-tag.md) - a non-element tag that calls a template with props from its attributes
* [Embedded statement](embedded-statement.md) - if, for, switch or select inside markup, recognized by the line rule
* [Interpolation](interpolation.md) - a {expr}: render call if it contains writer(), otherwise an escaped value
* [Attribute](attribute.md) - literal, boolean or dynamic attributes; props fields on component tags
* [writer()](writer-function.md) - the predeclared function returning the writer being rendered to
* [Golpe token](golpe-token.md) - go/token's token kinds extended with TEMPLATE and markup tokens

# Output
* [Element](element.md) - the renderable unit, func(io.Writer) error
* [Generated file](generated-file.md) - the committed x_gen.go produced from x.golpe
* [Source map](source-map.md) - line directives mapping generated Go back to .golpe positions
