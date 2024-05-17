package geecache

import (
	"cache/geecache/geecachepb"
	"cache/geecache/singleflight"
	"fmt"
	"log"
	"sync"
)

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name      string
	getter    Getter
	mainCache Cache
	peers     PeerPicker
	//singleflight 保证了fn只调用一次
	loader *singleflight.Group
}

func (g *Group) RegisterPeers(peer PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peer
}

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("getter nil")
	}
	mu.Lock()
	defer mu.Unlock()

	g := &Group{
		name:   name,
		getter: getter,
		mainCache: Cache{
			cacheBytes: cacheBytes,
		},
		loader: &singleflight.Group{},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	return groups[name]
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.Get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}
	return g.Load(key)
}

func (g *Group) Load(key string) (value ByteView, err error) {
	//if g.peers != nil {
	//	if peer, ok := g.peers.PeerPick(key); ok {
	//		if value, err := g.getFromPeer(key, peer); err == nil {
	//			return value, nil
	//		}
	//		log.Println("[GeeCache] Failed to get from peer", err)
	//	}
	//}
	view, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PeerPick(key); ok {
				if value, err = g.getFromPeer(key, peer); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}

		return g.GetLocally(key)
	})

	if err == nil {
		return view.(ByteView), nil
	}
	return
}

func (g *Group) getFromPeer(key string, peer PeerGetter) (value ByteView, err error) {
	//bytes, err := peer.Get(g.name, key)
	req := &geecachepb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &geecachepb.Response{}
	_, err = peer.Get(req, res)
	if err != nil {
		return ByteView{}, nil
	}
	return ByteView{b: res.Value}, nil
}

func (g *Group) GetLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{
		b: cloneBytes(bytes),
	}
	g.PopulateCache(key, value)
	return value, nil
}

func (g *Group) PopulateCache(key string, value ByteView) {
	g.mainCache.Put(key, value)
}
