module github.com/skaji/raku-cpan-new

go 1.14

require (
	github.com/dghubble/go-twitter v0.0.0-20190719072343-39e5462e111f
	github.com/dghubble/oauth1 v0.6.0
	github.com/skaji/go-nntp v0.0.2
	github.com/skaji/go-nntp-stream v0.0.0-20200607075831-0c87ee900b28
)

replace github.com/skaji/go-nntp-stream => /home/skaji/src/github.com/skaji/go-nntp-stream
