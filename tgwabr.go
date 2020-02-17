package tgwabr

import (
	"context"
	"log"
	appCtx "tgwabr/context"
	"tgwabr/pkg/cache"
	"tgwabr/pkg/store"
	"tgwabr/pkg/tg"
	"tgwabr/pkg/wa"
)

func Init() func() {
	var (
		err       error
		waImpl    *wa.Service
		tgImpl    *tg.Service
		storeImpl *store.Store
		cacheImpl *cache.Cache
	)

	ctx := context.Background()

	if storeImpl, err = store.New(ctx); err != nil {
		log.Fatalln("Fail Store Instance: ", err)
	}

	ctx = appCtx.NewDB(ctx, storeImpl)

	if waImpl, err = wa.New(ctx); err != nil {
		log.Fatalln("Fail WAInstance Instance: ", err)
	}

	ctx = appCtx.NewWA(ctx, waImpl)

	if tgImpl, err = tg.New(ctx); err != nil {
		log.Fatalln("Fail TG Instance: ", err)
	}

	ctx = appCtx.NewTG(ctx, tgImpl)

	if cacheImpl, err = cache.New(ctx, cache.Config{GetMembers: tgImpl.GetMembers}); err != nil {
		log.Fatalln("Fail Cache Instance: ", err)
	}

	ctx = appCtx.NewCache(ctx, cacheImpl)
	waImpl.UpdateCTX(ctx)

	return func() {
		if err = waImpl.ShutDown(); err != nil {
			log.Println("Fail WAInstance Instance: ", err)
		}
		if err = tgImpl.ShutDown(); err != nil {
			log.Println("Fail TG Instance: ", err)
		}
		if err = storeImpl.ShutDown(); err != nil {
			log.Println("Fail Store Instance: ", err)
		}
		if err = cacheImpl.ShutDown(); err != nil {
			log.Println("Fail Store Instance: ", err)
		}
	}
}
