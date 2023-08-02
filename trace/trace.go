package trace

import (
	"context"
	"fmt"
	"sync"

	"github.com/DistributedTraceProject/config"
)

const (
	Service              = "gaea"
	Jaeger  ProviderType = "jaeger"
)

type ProviderType string

var (
	providers       = make(map[ProviderType]Provider, 8)
	currentProvider Provider
	once            sync.Once
)

func RegisterProviders(pType ProviderType, p Provider) {
	providers[pType] = p
}

func Initialize(ctx context.Context, traceCfg *config.Trace) error {
	var err error
	once.Do(func() {
		v, ok := providers[ProviderType(traceCfg.Type)]
		if !ok {
			err = fmt.Errorf("not supported %s trace provider", traceCfg.Type)
			return
		}
		currentProvider = v
		err = currentProvider.Initialize(ctx, traceCfg)
	})
	return err
}

func Extract(ctx *config.Context, hints []*config.Hint) bool {
	return currentProvider.Extract(ctx, hints)
}

type Provider interface {
	Initialize(ctx context.Context, traceCfg *config.Trace) error
	Extract(ctx *config.Context, hints []*config.Hint) bool
}
