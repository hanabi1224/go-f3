package chainexchange_test

import (
	"context"
	"testing"
	"time"

	"github.com/filecoin-project/go-f3/chainexchange"
	"github.com/filecoin-project/go-f3/gpbft"
	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/stretchr/testify/require"
)

func TestPubSubChainExchange_Broadcast(t *testing.T) {
	const topicName = "fish"
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	var testInstant gpbft.Instant
	host, err := libp2p.New()
	require.NoError(t, err)
	t.Cleanup(func() {
		cancel()
		require.NoError(t, host.Close())
	})

	ps, err := pubsub.NewGossipSub(ctx, host, pubsub.WithFloodPublish(true))
	require.NoError(t, err)

	subject, err := chainexchange.NewPubSubChainExchange(
		chainexchange.WithProgress(func() (instant gpbft.Instant) {
			return testInstant
		}),
		chainexchange.WithPubSub(ps),
		chainexchange.WithTopicName(topicName),
		chainexchange.WithTopicScoreParams(nil),
	)
	require.NoError(t, err)
	require.NotNil(t, subject)

	err = subject.Start(ctx)
	require.NoError(t, err)

	instance := uint64(1)
	ecChain := gpbft.ECChain{
		{Epoch: 0, Key: []byte("lobster"), PowerTable: gpbft.MakeCid([]byte("pt"))},
		{Epoch: 1, Key: []byte("barreleye"), PowerTable: gpbft.MakeCid([]byte("pt"))},
	}

	key := subject.Key(ecChain)
	chain, found := subject.GetChainByInstance(ctx, instance, key)
	require.False(t, found)
	require.Nil(t, chain)

	require.NoError(t, subject.Broadcast(ctx, chainexchange.Message{
		Instance: instance,
		Chain:    ecChain,
	}))

	chain, found = subject.GetChainByInstance(ctx, instance, key)
	require.True(t, found)
	require.Equal(t, ecChain, chain)

	baseChain := ecChain.BaseChain()
	baseKey := subject.Key(baseChain)
	chain, found = subject.GetChainByInstance(ctx, instance, baseKey)
	require.True(t, found)
	require.Equal(t, baseChain, chain)

	require.NoError(t, subject.Shutdown(ctx))
}

// TODO: Add more tests, specifically:
//        - valodation
//        - discovery through other chainexchange instance
//        - cache eviction/fixed memory footprint.
//        - fulfilment of chain from discovery to wanted in any order.
//        - spam
//        - fuzz
