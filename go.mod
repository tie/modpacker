module github.com/tie/modpacker

go 1.13

require (
	github.com/andybalholm/cascadia v1.1.0
	github.com/google/subcommands v1.0.1
	github.com/hashicorp/hcl/v2 v2.1.0
	github.com/pkg/diff v0.0.0-20190930165518-531926345625
	github.com/stretchr/testify v1.4.0 // indirect
	github.com/tie/internal v0.0.0-20191125222958-4c3152d9f9ef
	github.com/zclconf/go-cty v1.1.0
	golang.org/x/crypto v0.0.0-20190510104115-cbcb75029529
	golang.org/x/net v0.0.0-20191119073136-fc4aabc6c914
	golang.org/x/sys v0.0.0-20191008105621-543471e840be // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/src-d/go-billy.v4 v4.3.2
)

replace gopkg.in/src-d/go-billy.v4 v4.3.2 => github.com/tie/go-billy v0.0.0-20191123113025-16c87c3285a3
