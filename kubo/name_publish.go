package kubo

import (
	"context"
	"fmt"

	"github.com/NatoBoram/ipapm/wheel"
	"github.com/ipfs/boxo/ipns"
	"github.com/ipfs/kubo/core/commands/cmdutils"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/ipfs/kubo/core/coreiface/options"
)

func (k *Client) NamePublish(ctx context.Context, keyName, cid string) (ipns.Name, error) {
	kkey := k.Key()
	keys, err := kkey.List(ctx)
	if err != nil {
		return ipns.Name{}, fmt.Errorf("listing keys: %w", err)
	}

	_, ok := wheel.Find(keys, func(key iface.Key) bool {
		return key.Name() == keyName
	})
	if !ok {
		_, err := kkey.Generate(ctx, keyName)
		if err != nil {
			return ipns.Name{}, fmt.Errorf("generating key %q: %w", keyName, err)
		}
	}

	p, err := cmdutils.PathOrCidPath(cid)
	if err != nil {
		return ipns.Name{}, fmt.Errorf("parsing CID %q: %w", cid, err)
	}

	opt := options.Name.Key(keyName)
	name, err := k.Name().Publish(ctx, p, opt)
	if err != nil {
		return ipns.Name{}, fmt.Errorf("publishing %q to IPNS: %w", cid, err)
	}

	return name, nil
}
