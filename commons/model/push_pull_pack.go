package model

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// PushPullBitXXX denotes a bit for the option for PushPull
const (
	PushPullBitNormal      PushPullPackOption = 0x00
	PushPullBitCreate      PushPullPackOption = 0x01
	PushPullBitSubscribe   PushPullPackOption = 0x02
	PushPullBitUnsubscribe PushPullPackOption = 0x04
	PushPullBitDelete      PushPullPackOption = 0x08
	PushPullBitSnapshot    PushPullPackOption = 0x10
	PushPullBitError       PushPullPackOption = 0x20
)

var pushPullBitString = []string{"cr", "sb", "un", "de", "sn", "er"}

type PushPullPackOption uint32

func (p *PushPullPackOption) String() string {
	var bit = uint32(*p)
	var ret = "[ "
	for i := 0; i < len(pushPullBitString); i++ {
		b := bit & 0x01
		if b == 0 {
			ret +=
				pushPullBitString[i] + " "
		} else {
			ret += strings.ToUpper(pushPullBitString[i]) + " "
		}
		bit = bit >> 1
	}
	return ret + "]"
}

// SetCreateBit sets CreateBit
func (p *PushPullPackOption) SetCreateBit() *PushPullPackOption {
	*p |= PushPullBitCreate
	return p
}

// SetSubscribeBit sets SubscribeBit
func (p *PushPullPackOption) SetSubscribeBit() *PushPullPackOption {
	*p |= PushPullBitSubscribe
	return p
}

// SetUnsubscribeBit sets UnsubscribeBit
func (p *PushPullPackOption) SetUnsubscribeBit() *PushPullPackOption {
	*p |= PushPullBitUnsubscribe
	return p
}

// SetDeleteBit sets DeleteBit
func (p *PushPullPackOption) SetDeleteBit() *PushPullPackOption {
	*p |= PushPullBitDelete
	return p
}

// SetSnapshotBit sets SnapshotBit
func (p *PushPullPackOption) SetSnapshotBit() *PushPullPackOption {
	*p |= PushPullBitSnapshot
	return p
}

// SetErrorBit sets ErrorBit
func (p *PushPullPackOption) SetErrorBit() *PushPullPackOption {
	*p |= PushPullBitError
	return p
}

// HasCreateBit examines CreateBit
func (p *PushPullPackOption) HasCreateBit() bool {
	return (*p & PushPullBitCreate) == PushPullBitCreate
}

// HasSubscribeBit examines SubscribeBit
func (p *PushPullPackOption) HasSubscribeBit() bool {
	return (*p & PushPullBitSubscribe) == PushPullBitSubscribe
}

// HasUnsubscribeBit examines UnsubscribeBit
func (p *PushPullPackOption) HasUnsubscribeBit() bool {
	return (*p & PushPullBitUnsubscribe) == PushPullBitUnsubscribe
}

// HasDeleteBit examines DeleteBit
func (p *PushPullPackOption) HasDeleteBit() bool {
	return (*p & PushPullBitDelete) == PushPullBitDelete
}

// HasSnapshotBit examines SnapshotBit
func (p *PushPullPackOption) HasSnapshotBit() bool {
	return (*p & PushPullBitSnapshot) == PushPullBitSnapshot
}

// HasErrorBit examines ErrorBit
func (p *PushPullPackOption) HasErrorBit() bool {
	return (*p & PushPullBitError) == PushPullBitError
}

// GetPushPullPackOption returns PushPullOption
func (p *PushPullPack) GetPushPullPackOption() *PushPullPackOption {
	var option = (*PushPullPackOption)(&p.Option)
	return option
}

// GetResponsePushPullPack returns the PushPullPack that can be used for response.
func (p *PushPullPack) GetResponsePushPullPack() *PushPullPack {
	return &PushPullPack{
		Key:        p.Key,
		DUID:       p.DUID,
		Option:     p.Option,
		Era:        p.Era,
		Type:       p.Type,
		CheckPoint: p.CheckPoint.Clone(),
	}
}

// ToString returns customized string
func (p *PushPullPack) ToString() string {
	var b strings.Builder
	var option = PushPullPackOption(p.Option)

	_, _ = fmt.Fprintf(&b, "%s(%s) %s CP(%v) OP(%d){", p.Key, hex.EncodeToString(p.DUID), option.String(), p.CheckPoint.String(), len(p.Operations))
	for _, op := range p.Operations {
		b.WriteString(op.ToString())
		b.WriteString(" =>")
	}
	b.WriteString("}")
	return b.String()
}
