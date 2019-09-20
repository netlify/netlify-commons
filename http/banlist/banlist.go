package banlist

import (
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Domains []string
	URLs    []string
}

type Banlist struct {
	domainHolder atomic.Value
	urlHolder    atomic.Value
	mtx          sync.Mutex
	ch           chan os.Signal
	log          logrus.FieldLogger
	path         string
}

func New(log logrus.FieldLogger, path string) *Banlist {
	bl := newBanlist(log, path)
	bl.listen()
	bl.runUpdate()
	return bl
}

func newBanlist(log logrus.FieldLogger, path string) *Banlist {
	bl := &Banlist{log: log, path: path}
	bl.domainHolder.Store(make(map[string]struct{}))
	bl.urlHolder.Store(make(map[string]struct{}))
	return bl
}

func (b *Banlist) listen() {
	b.ch = make(chan os.Signal, 1)
	signal.Notify(b.ch, syscall.SIGHUP)
	go func() {
		for range b.ch {
			b.runUpdate()
		}
		b.log.Info("No longer listening for SIGHUP")
	}()
}

func (b *Banlist) runUpdate() {
	if err := b.update(); err != nil {
		b.log.WithError(err).Warn("error updating banlist")
	} else {
		b.log.Info("banlist updated")
	}
}

func (b *Banlist) update() error {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	f, err := os.Open(b.path)
	if err != nil {
		return errors.Wrap(err, "error opening banlist config")
	}
	defer f.Close()

	c := new(Config)
	if err := json.NewDecoder(f).Decode(c); err != nil {
		return errors.Wrap(err, "error decoding banlist config")
	}

	domains := make(map[string]struct{})
	urls := make(map[string]struct{})
	for _, el := range c.Domains {
		domains[strings.ToLower(el)] = struct{}{}
	}
	for _, el := range c.URLs {
		urls[strings.ToLower(el)] = struct{}{}
	}

	b.domainHolder.Store(domains)
	b.urlHolder.Store(urls)
	return nil
}

// CheckRequest will check if the domain is blocked or the path is blocked
func (b *Banlist) CheckRequest(r *http.Request) bool {
	domain := strings.SplitN(r.Host, ":", 2)[0]
	if _, ok := b.domains()[strings.ToLower(domain)]; ok {
		return true
	}

	url := domain + r.URL.Path
	if _, ok := b.urls()[strings.ToLower(url)]; ok {
		return true
	}

	return false
}

func (b *Banlist) Close() {
	signal.Stop(b.ch)
	close(b.ch)
}

func (b *Banlist) domains() map[string]struct{} {
	return b.domainHolder.Load().(map[string]struct{})
}

func (b *Banlist) urls() map[string]struct{} {
	return b.urlHolder.Load().(map[string]struct{})
}
