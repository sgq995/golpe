# golpe

A Golang web framework built on **Golpe**, a meta-language over Go for
writing HTML views, in the spirit of [templ](https://templ.guide) and Ripple.

```golpe
type NavItemProps struct {
    Route    string
    Selected bool
    Children element
}

template NavItem(props NavItemProps) {
    <li>
    if props.Selected {
        <span>{props.Children}</span>
    } else {
        <a href={props.Route}>{props.Children}</a>
    }
    </li>
}
```

`golpe generate` turns each `x.golpe` into `x_golpe.go`: plain Go where
`NavItem` is a `func(NavItemProps) golpe.Element` and
`golpe.Element = func(io.Writer) error`. Values are HTML-escaped and every
write error is propagated.

> **Status:** design phase. The language, generator, CLI and LSP are
> specified; implementation is starting. See the
> [knowledge base](docs/knowledge/index.md).

## Project knowledge

Specs (what and why), plans (how), task slices (todo lists), decisions,
actors, actuators and entities live in [`docs/knowledge`](docs/knowledge/index.md)
as an [Open Knowledge Format](https://okf.md) bundle. Features are described
in chat; agents write them up following
[the feature workflow](docs/knowledge/process/feature-workflow.md).

## Development

```sh
go build -o golpe ./cmd/cli
gofmt -l . && go vet ./... && go test ./...
go run ./internal/okflint      # validate the knowledge base
```
