package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"os"
	"time"

	"github.com/rs/xid"

	"github.com/brick-io/brock/sdk"
	sdkotel "github.com/brick-io/brock/sdk/otel"
)

/*
blockchain is basically a ledger (just like in the bank)
but rather than having a centralized authority
the information is spread across independent parties
to minimize any fraud


start with genesis the first blockchain_like info,
then the first tx usually marking an ICO (initial coin offering)
*/

//nolint:gochecknoglobals
var (
	_GOD = &Account{id: xid.New().Bytes(), name: "_GOD"}

	ErrInsufficientFund = sdk.Errorf("blockchain: insufficient fund")
)

func main() {
	ctx := context.Background()
	log := sdkotel.Log(ctx, os.Stdout)

	theBlock := &Block{id: xid.New().Bytes()}

	var err error

	Alex := &Account{id: xid.New().Bytes(), name: "Alex"}
	Bayu := &Account{id: xid.New().Bytes(), name: "Bayu"}
	Chad := &Account{id: xid.New().Bytes(), name: "Chad"}
	format := "%+v, %v\n\n"

	theBlock, err = theBlock.Chain(&Record{
		id: xid.New().Bytes(), ts: time.Now(),
		amt: (100), src: _GOD, dst: Alex,
	})
	log.Log.Print(format, theBlock.record, err)

	theBlock, err = theBlock.Chain(&Record{
		id: xid.New().Bytes(), ts: time.Now(),
		amt: (100), src: _GOD, dst: Bayu,
	})
	log.Log.Print(format, theBlock.record, err)

	theBlock, err = theBlock.Chain(&Record{
		id: xid.New().Bytes(), ts: time.Now(),
		amt: (35), src: Alex, dst: Bayu,
	})
	log.Log.Print(format, theBlock.record, err)

	theBlock, err = theBlock.Chain(&Record{
		id: xid.New().Bytes(), ts: time.Now(),
		amt: (35), src: Alex, dst: Bayu,
	})
	log.Log.Print(format, theBlock.record, err)

	theBlock, err = theBlock.Chain(&Record{
		id: xid.New().Bytes(), ts: time.Now(),
		amt: (35), src: Alex, dst: Bayu,
	})
	log.Log.Print(format, theBlock.record, err)

	theBlock, err = theBlock.Chain(&Record{
		id: xid.New().Bytes(), ts: time.Now(),
		amt: (35), src: Bayu, dst: Chad,
	})
	log.Log.Print(format, theBlock.record, err)

	theBlock, err = theBlock.Chain(&Record{
		id: xid.New().Bytes(), ts: time.Now(),
		amt: (35), src: Bayu, dst: Chad,
	})
	log.Log.Print(format, theBlock.record, err)
}

// Account
// contains Account info.
type Account struct {
	id   []byte
	name string
}

func (a *Account) String() string {
	id, _ := xid.FromBytes(a.id)

	return "[..." + last((3), id.String()) + " " + a.name + "]"
}

// Record
// from src to dst, amassing some amt
// srcBalance & dstBalance snapshot the last balance after a Record is written
// providing fast lookup on the last balance.
type Record struct {
	id []byte

	ts  time.Time
	amt uint
	src *Account
	dst *Account

	srcLastBalance uint // generate by Chain
	dstLastBalance uint // generate by Chain
}

func (r *Record) String() string {
	srcLastBalance := sdk.Sprint(r.srcLastBalance)
	if r.src == _GOD {
		srcLastBalance = "♾️"
	}

	return sdk.Sprintf(
		"%x\n[%s] %s -> %s (%d) %d -> %s",
		r.Hash(),
		r.ts.Format(time.RFC3339),
		r.src,
		srcLastBalance,
		r.amt,
		r.dstLastBalance,
		r.dst,
	)
}

func (r *Record) Hash() []byte {
	data := []byte(r.ts.Format(time.RFC3339Nano))
	data = append(data, r.src.id...)
	data = append(data, []byte(sdk.Sprint(r.amt))...)
	data = append(data, r.dst.id...)
	h := sha256.Sum256(data)
	r.id = h[:]

	return r.id
}

// Block
// is a unit of distributed ledger, wrapping the record.
type Block struct {
	id []byte

	parent *Block
	record *Record

	srcLastBlock *Block // generate by Chain
	dstLastBlock *Block // generate by Chain
}

// Chain
// new block to the current block, then return the new block.
func (b *Block) Chain(record *Record) (*Block, error) {
	var (
		srcLastBlock, dstLastBlock *Block
		srcBalance, dstBalance     uint
	)

	if i := 0; record.src == _GOD {
		srcBalance += record.amt // normalize to 0
		srcLastBlock = new(Block)
		dstLastBlock = new(Block)
	} else {
		srcLastBlock, i = b.LastBlockOf(record.src)
		srcBalance = balanceOfLastBlock(srcLastBlock, i)
		dstLastBlock, i = b.LastBlockOf(record.dst)
		dstBalance = balanceOfLastBlock(dstLastBlock, i)
		if srcBalance < record.amt {
			return b, ErrInsufficientFund
		}
	}

	record.srcLastBalance = srcBalance - record.amt
	record.dstLastBalance = dstBalance + record.amt

	return &Block{
		id:           record.Hash(),
		srcLastBlock: srcLastBlock,
		dstLastBlock: dstLastBlock,
		parent:       b,
		record:       record,
	}, nil
}

func balanceOfLastBlock(b *Block, i int) uint {
	if i != 0 && b != nil {
		if i < 0 {
			return b.record.dstLastBalance
		}

		return b.record.srcLastBalance
	}

	return 0
}

func (b *Block) LastBlockOf(a *Account) (*Block, int) {
	if b.parent == nil {
		return nil, 0
	} else if b.record != nil && bytes.Equal(a.id, b.record.src.id) {
		return b, 1
	} else if b.record != nil && bytes.Equal(a.id, b.record.dst.id) {
		return b, -1
	}

	return b.parent.LastBlockOf(a)
}

func last(n int, s string) string {
	return s[len(s)-n:]
}
