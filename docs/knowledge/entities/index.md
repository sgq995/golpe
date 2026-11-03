# Entities
The things Golpe reads, produces and reasons about.

# Source
* [Golpe source file](golpe-file.md) - a .golpe file: Go that may also contain template declarations
* [Template declaration](template-declaration.md) - a top-level template whose body mixes Go and markup and yields an Element
* [Markup node](markup-node.md) - a node of the HTML tree inside a template
* [Embedded statement](embedded-statement.md) - if, for, switch or select inside markup, with markup-only bodies
* [Interpolation](interpolation.md) - a {expr} Go expression inside markup or an attribute
* [Attribute](attribute.md) - a name or name=value pair on a tag
* [Golpe token](golpe-token.md) - go/token's token kinds extended with TEMPLATE and markup tokens

# Output
* [Element](element.md) - the renderable unit, func(io.Writer) error
* [Generated file](generated-file.md) - the x_golpe.go file produced from x.golpe
* [Source map](source-map.md) - line directives mapping generated Go back to .golpe positions
