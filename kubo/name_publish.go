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
		return ipns.Name{}, fmt.Errorf("failed to list keys: %w", err)
	}

	_, ok := wheel.Find(keys, func(key iface.Key) bool {
		return key.Name() == keyName
	})
	if !ok {
		_, err := kkey.Generate(ctx, keyName)
		if err != nil {
			return ipns.Name{}, fmt.Errorf("failed to generate key %s: %w", keyName, err)
		}
	}

	p, err := cmdutils.PathOrCidPath(cid)
	if err != nil {
		return ipns.Name{}, fmt.Errorf("invalid CID: %w", err)
	}

	opt := options.Name.Key(keyName)
	name, err := k.Name().Publish(ctx, p, opt)
	if err != nil {
		return ipns.Name{}, fmt.Errorf("failed to publish to IPNS: %w", err)
	}

	return name, nil
}
