package model

import (
	"fmt"
	"github.com/orda-io/orda/client/pkg/utils"
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
	PushPullBitReadOnly    PushPullPackOption = 0x40
)

var pushPullBitString = []string{"cr", "sb", "un", "de", "sn", "er", "ro"}

// PushPullPackOption denotes an option implied in a PushPullPack.
type PushPullPackOption uint32

func (its *PushPullPackOption) String() string {
	var bit = uint32(*its)
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

// SetCreateBit sets CreateBit.
func (its *PushPullPackOption) SetCreateBit() *PushPullPackOption {
	*its |= PushPullBitCreate
	return its
}

// SetSubscribeBit sets SubscribeBit.
func (its *PushPullPackOption) SetSubscribeBit() *PushPullPackOption {
	*its |= PushPullBitSubscribe
	return its
}

// SetUnsubscribeBit sets UnsubscribeBit.
func (its *PushPullPackOption) SetUnsubscribeBit() *PushPullPackOption {
	*its |= PushPullBitUnsubscribe
	return its
}

// SetDeleteBit sets DeleteBit.
func (its *PushPullPackOption) SetDeleteBit() *PushPullPackOption {
	*its |= PushPullBitDelete
	return its
}

// SetSnapshotBit sets SnapshotBit.
func (its *PushPullPackOption) SetSnapshotBit() *PushPullPackOption {
	*its |= PushPullBitSnapshot
	return its
}

// SetErrorBit sets ErrorBit.
func (its *PushPullPackOption) SetErrorBit() *PushPullPackOption {
	*its |= PushPullBitError
	return its
}

// SetReadOnlyBit sets ReadOnlyBit
func (its *PushPullPackOption) SetReadOnlyBit() *PushPullPackOption {
	*its |= PushPullBitReadOnly
	return its
}

// HasCreateBit examines CreateBit.
func (its *PushPullPackOption) HasCreateBit() bool {
	return (*its & PushPullBitCreate) == PushPullBitCreate
}

// HasSubscribeBit examines SubscribeBit.
func (its *PushPullPackOption) HasSubscribeBit() bool {
	return (*its & PushPullBitSubscribe) == PushPullBitSubscribe
}

// HasUnsubscribeBit examines UnsubscribeBit.
func (its *PushPullPackOption) HasUnsubscribeBit() bool {
	return (*its & PushPullBitUnsubscribe) == PushPullBitUnsubscribe
}

// HasDeleteBit examines DeleteBit.
func (its *PushPullPackOption) HasDeleteBit() bool {
	return (*its & PushPullBitDelete) == PushPullBitDelete
}

// HasSnapshotBit examines SnapshotBit.
func (its *PushPullPackOption) HasSnapshotBit() bool {
	return (*its & PushPullBitSnapshot) == PushPullBitSnapshot
}

// HasErrorBit examines ErrorBit.
func (its *PushPullPackOption) HasErrorBit() bool {
	return (*its & PushPullBitError) == PushPullBitError
}

// HasReadOnly examines ReadOnlyBit.
func (its *PushPullPackOption) HasReadOnly() bool {
	return (*its & PushPullBitReadOnly) == PushPullBitReadOnly
}

// GetPushPullPackOption returns PushPullOption.
func (its *PushPullPack) GetPushPullPackOption() *PushPullPackOption {
	var option = (*PushPullPackOption)(&its.Option)
	return option
}

// GetDatatypeTag returns datatype tag
func (its *PushPullPack) GetDatatypeTag() string {
	return fmt.Sprintf("%s(%s)", utils.MakeDefaultShort(its.Key), its.DUID)
}

// GetResponsePushPullPack returns the PushPullPack that can be used for response.
func (its *PushPullPack) GetResponsePushPullPack() *PushPullPack {
	return &PushPullPack{
		Key:        its.Key,
		DUID:       its.DUID,
		Option:     its.Option,
		Era:        its.Era,
		Type:       its.Type,
		CheckPoint: its.CheckPoint.Clone(),
	}
}

// ToString returns customized string.
func (its *PushPullPack) ToString(isFull bool) string {
	var option = PushPullPackOption(its.Option)
	var opList = OpList(its.Operations)
	return fmt.Sprintf(
		"%s %s(%s) %s CP%s OP(%d){%v}",
		its.Type,
		its.Key,
		its.DUID,
		option.String(),
		its.CheckPoint.ToString(),
		len(its.Operations),
		opList.ToString(isFull))
}
