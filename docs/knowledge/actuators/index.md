# Actuators
Golpe's own components: the parts that perform work.

* [Golpe CLI](golpe-cli.md) - the golpe command that generates Go and hosts the language server
* [Parser](parser.md) - turns a .golpe file into a Go AST extended with template and markup nodes
* [Generator](generator.md) - lowers the extended AST into plain, gofmt-ed Go
* [Runtime](runtime.md) - the public root package: Element and the write/escape helpers
* [Language server](language-server.md) - proxies editor requests to gopls and maps positions back
* [okflint](okflint.md) - validates this knowledge base
