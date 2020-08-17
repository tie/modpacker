module github.com/tie/modpacker

go 1.14

require (
	github.com/andybalholm/cascadia v1.2.0
	github.com/go-git/go-billy/v5 v5.0.0
	github.com/google/subcommands v1.2.0
	github.com/hashicorp/hcl/v2 v2.6.0
	github.com/pkg/diff v0.0.0-20190930165518-531926345625
	github.com/tie/internal v0.0.0-20191125222958-4c3152d9f9ef
	github.com/zclconf/go-cty v1.5.1
	golang.org/x/crypto v0.0.0-20200728195943-123391ffb6de
	golang.org/x/net v0.0.0-20200813134508-3edf25e44fcc
)

// TODO remove once https://github.com/go-git/go-billy/pull/7 is merged
replace github.com/go-git/go-billy/v5 => github.com/tie/go-billy/v5 v5.0.1-0.20200817232414-4055a2947b21
