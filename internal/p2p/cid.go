package p2p

import (
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

func createCID(key []byte) (cid.Cid, error) {
	mh, err := multihash.Sum(key, multihash.SHA2_256, -1)
	if err != nil {
		return cid.Undef, err
	}

	return cid.NewCidV1(cid.Raw, mh), nil
}
