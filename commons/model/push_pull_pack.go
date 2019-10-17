package model

const (
	PushPullBitNormal      PushPullPackOption = 0x00
	PushPullBitCreate      PushPullPackOption = 0x01
	PushPullBitSubscribe   PushPullPackOption = 0x02
	PushPullBitUnsubscribe PushPullPackOption = 0x04
	PushPullBitDelete      PushPullPackOption = 0x08
	PushPullBitSnapshot    PushPullPackOption = 0x10
	PushPullBitError       PushPullPackOption = 0x20
)

type PushPullPackOption uint32

func (p PushPullPackOption) SetNormalBit() PushPullPackOption {
	p |= PushPullBitNormal
	return p
}

func (p PushPullPackOption) SetCreateBit() PushPullPackOption {
	p |= PushPullBitCreate
	return p
}

func (p PushPullPackOption) SetSubscribeBit() PushPullPackOption {
	p |= PushPullBitSubscribe
	return p
}

func (p PushPullPackOption) SetUnsubscribeBit() PushPullPackOption {
	p |= PushPullBitUnsubscribe
	return p
}

func (p PushPullPackOption) SetDeleteBit() PushPullPackOption {
	p |= PushPullBitDelete
	return p
}

func (p PushPullPackOption) SetSnapshotBit() PushPullPackOption {
	p |= PushPullBitSnapshot
	return p
}

func (p PushPullPackOption) SetErrorBit() PushPullPackOption {
	p |= PushPullBitError
	return p
}

func (p PushPullPackOption) HasCreateBit() bool {
	return (p & PushPullBitCreate) == PushPullBitCreate
}

func (p PushPullPackOption) HasSubscribeBit() bool {
	return (p & PushPullBitSubscribe) == PushPullBitSubscribe
}

func (p PushPullPackOption) HasUnsubscribeBit() bool {
	return (p & PushPullBitUnsubscribe) == PushPullBitUnsubscribe
}

func (p PushPullPackOption) HasDeleteBit() bool {
	return (p & PushPullBitDelete) == PushPullBitDelete
}

func (p PushPullPackOption) HasSnapshotBit() bool {
	return (p & PushPullBitSnapshot) == PushPullBitSnapshot
}

func (p PushPullPackOption) HasErrorBit() bool {
	return (p & PushPullBitError) == PushPullBitError
}

func (p *PushPullPack) GetPushPullPackOption() PushPullPackOption {
	return PushPullPackOption(p.Option)
}
