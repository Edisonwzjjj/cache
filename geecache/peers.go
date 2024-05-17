package geecache

import pb "cache/geecache/geecachepb"

type PeerPicker interface {
	PeerPick(key string) (peer PeerGetter, ok bool)
}

type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error
}
