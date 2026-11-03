# golpe

**Golpe** is a meta-language over Go for writing server-rendered HTML,
borrowing part of the syntax of [templ](https://templ.guide) and Ripple.
`.golpe` files are Go plus `template` declarations whose bodies switch into
HTML markup.

```golpe
type NavItemProps struct {
    Route    string
    Selected bool
    Children element
}

template NavItem(props NavItemProps) {
    <li>
    if props.Selected {
        <span>{props.Children(writer())}</span>
    } else {
        <a href={props.Route}>{props.Children(writer())}</a>
    }
    </li>
}
```

`golpe generate` turns each `x.golpe` into a committed `x_gen.go`: plain Go
where `NavItem` is a `func(NavItemProps) func(io.Writer) error`. The generated
code imports only the standard library (`fmt`, `html`, `io`): values are
HTML-escaped, child elements are rendered explicitly with `writer()`, and
every write error is returned with its `.golpe` position.

> **Status:** design and early implementation. The language, generator, CLI,
> component tags and LSP are specified; the token and AST packages are
> built. See the [knowledge base](docs/knowledge/index.md).

## Install

```sh
go install github.com/sgq995/golpe/cmd/golpe@latest
# or, as a module tool:
go get -tool github.com/sgq995/golpe/cmd/golpe && go tool golpe
```

The CLI is the only thing a project needs; the root package
`github.com/sgq995/golpe` holds optional helpers.

## Project knowledge

Specs (what and why), plans (how), task slices (todo lists), decisions,
actors, actuators, entities and a backlog live in
[`docs/knowledge`](docs/knowledge/index.md) as an
[Open Knowledge Format](https://okf.md) bundle. Features are described in
chat; agents write them up following
[the feature workflow](docs/knowledge/process/feature-workflow.md).

## Development

```sh
go build ./cmd/golpe
gofmt -l . && go vet ./... && go test ./...
go run ./internal/okflint      # validate the knowledge base
```

## License

[MIT](LICENSE)
