package stream

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/skaji/go-nntp-stream"
	"github.com/skaji/raku-cpan-new/pkg/distribution"
	"github.com/skaji/raku-cpan-new/pkg/log"
)

func fixRakuDistribution(ctx context.Context, d *distribution.Distribution) error {
	fetcher := distribution.NewRakuFetcher()
	max := 20
	for i := 1; i <= max; i++ {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		name, err := fetcher.FetchName(ctx, d.MetaURL())
		cancel()
		if err == nil {
			if d.MainModule == name {
				log.Printf("%d guessed MainModule %s matches name in META file", d.ID, d.MainModule)
			} else {
				log.Printf("%d guessed MainModule %s does NOT match name (%s) in META file, use the name in META file", d.ID, d.MainModule, name)
				d.MainModule = name
			}
			return nil
		}
		if _, ok := err.(*distribution.RetryableError); !ok {
			return err
		}

		log.Print(d.ID, err)
		if i != max {
			log.Print(d.ID, "Sleep 30sec...")
			time.Sleep(30 * time.Second)
		}
	}
	return errors.New("too many retry, give up")
}

func NewRaku(ctx context.Context, addr string, tick time.Duration) <-chan *distribution.Distribution {
	ch := make(chan *distribution.Distribution)
	go func() {
		defer close(ch)

		stream := nntp.Stream(ctx,
			nntp.WithAddr(addr),
			nntp.WithGroup("perl.cpan.uploads"),
			nntp.WithTick(tick),
			nntp.WithTimeout(25*time.Second),
			nntp.WithSubscribeLog(true),
		)

		seen := map[int]struct{}{}
		for event := range stream {
			if l, ok := event.(*nntp.Log); ok {
				if l.Level == nntp.LogLevelDebug {
					log.Debug(l.Message)
				} else {
					log.Print(l.Message)
				}
				continue
			}
			article := event.(*nntp.Article)
			id := article.ID
			subject := article.Header.Get("Subject")
			dist, err := distribution.New(id, subject)
			if err != nil {
				log.Print(id, err)
				continue
			}

			log.Print(id, dist.AsJSON())
			if !dist.IsRaku {
				continue
			}
			if _, ok := seen[id]; ok {
				log.Print(id, fmt.Sprintf("Already seen %d, skip", id))
				continue
			}
			seen[id] = struct{}{}

			go func(id int) {
				err := fixRakuDistribution(ctx, dist)
				if err == nil {
					ch <- dist
				} else {
					log.Print(id, err)
				}
			}(id)
		}
	}()
	return ch
}
