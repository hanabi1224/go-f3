package f3

import (
	"math/rand"
	"testing"

	"github.com/filecoin-project/go-bitfield"
	"github.com/filecoin-project/go-f3/gpbft"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/require"
)

const seed = 1413

func BenchmarkCborEncoding(b *testing.B) {
	rng := rand.New(rand.NewSource(seed))
	encoder := &cborGMessageEncoding{}
	msg := generateRandomPartialGMessage(b, rng)

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := encoder.Encode(msg); err != nil {
				require.NoError(b, err)
			}
		}
	})
}

func BenchmarkCborDecoding(b *testing.B) {
	rng := rand.New(rand.NewSource(seed))
	encoder := &cborGMessageEncoding{}
	msg := generateRandomPartialGMessage(b, rng)
	data, err := encoder.Encode(msg)
	require.NoError(b, err)

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if got, err := encoder.Decode(data); err != nil {
				require.NoError(b, err)
				require.Equal(b, msg, got)
			}
		}
	})
}

func BenchmarkZstdEncoding(b *testing.B) {
	rng := rand.New(rand.NewSource(seed))
	encoder, err := newZstdGMessageEncoding()
	require.NoError(b, err)
	msg := generateRandomPartialGMessage(b, rng)

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := encoder.Encode(msg); err != nil {
				require.NoError(b, err)
			}
		}
	})
}

func BenchmarkZstdDecoding(b *testing.B) {
	rng := rand.New(rand.NewSource(seed))
	encoder, err := newZstdGMessageEncoding()
	require.NoError(b, err)
	msg := generateRandomPartialGMessage(b, rng)
	data, err := encoder.Encode(msg)
	require.NoError(b, err)

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if got, err := encoder.Decode(data); err != nil {
				require.NoError(b, err)
				require.Equal(b, msg, got)
			}
		}
	})
}

func generateRandomPartialGMessage(b *testing.B, rng *rand.Rand) *PartialGMessage {
	var pgmsg PartialGMessage
	pgmsg.GMessage = generateRandomGMessage(b, rng)
	pgmsg.GMessage.Vote.Value = nil
	if pgmsg.Justification != nil {
		pgmsg.GMessage.Justification.Vote.Value = nil
	}
	pgmsg.VoteValueKey = generateRandomBytes(b, rng, 32)
	return &pgmsg
}

func generateRandomGMessage(b *testing.B, rng *rand.Rand) *gpbft.GMessage {
	var maybeTicket []byte
	if rng.Float64() < 0.5 {
		generateRandomBytes(b, rng, 96)
	}

	return &gpbft.GMessage{
		Sender:        gpbft.ActorID(rng.Uint64()),
		Vote:          generateRandomPayload(b, rng),
		Signature:     generateRandomBytes(b, rng, 96),
		Ticket:        maybeTicket,
		Justification: generateRandomJustification(b, rng),
	}
}

func generateRandomJustification(b *testing.B, rng *rand.Rand) *gpbft.Justification {
	return &gpbft.Justification{
		Vote:      generateRandomPayload(b, rng),
		Signers:   generateRandomBitfield(rng),
		Signature: generateRandomBytes(b, rng, 96),
	}
}

func generateRandomBytes(b *testing.B, rng *rand.Rand, n int) []byte {
	buf := make([]byte, n)
	_, err := rng.Read(buf)
	require.NoError(b, err)
	return buf
}

func generateRandomPayload(b *testing.B, rng *rand.Rand) gpbft.Payload {
	return gpbft.Payload{
		Instance: rng.Uint64(),
		Round:    rng.Uint64(),
		Phase:    gpbft.Phase(rng.Intn(int(gpbft.COMMIT_PHASE)) + 1),
		Value:    generateRandomECChain(b, rng, rng.Intn(gpbft.ChainMaxLen)+1),
		SupplementalData: gpbft.SupplementalData{
			PowerTable: generateRandomCID(b, rng),
		},
	}
}

func generateRandomBitfield(rng *rand.Rand) bitfield.BitField {
	ids := make([]uint64, rng.Intn(2_000)+1)
	for i := range ids {
		ids[i] = rng.Uint64()
	}
	return bitfield.NewFromSet(ids)
}

func generateRandomECChain(b *testing.B, rng *rand.Rand, length int) gpbft.ECChain {
	chain := make(gpbft.ECChain, length)
	epoch := int64(rng.Uint64())
	for i := range length {
		chain[i] = generateRandomTipSet(b, rng, epoch+int64(i))
	}
	return chain
}

func generateRandomTipSet(b *testing.B, rng *rand.Rand, epoch int64) gpbft.TipSet {
	return gpbft.TipSet{
		Epoch:      epoch,
		Key:        generateRandomTipSetKey(b, rng),
		PowerTable: generateRandomCID(b, rng),
	}
}

func generateRandomTipSetKey(b *testing.B, rng *rand.Rand) gpbft.TipSetKey {
	key := make([]byte, rng.Intn(gpbft.TipsetKeyMaxLen)+1)
	_, err := rng.Read(key)
	require.NoError(b, err)
	return key
}

func generateRandomCID(b *testing.B, rng *rand.Rand) cid.Cid {
	sum, err := multihash.Sum(generateRandomBytes(b, rng, 32), multihash.SHA2_256, -1)
	require.NoError(b, err)
	return cid.NewCidV1(cid.Raw, sum)
}
