package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgRegisterSidecars{}
	_ sdk.Msg = &MsgHeartbeat{}
	_ sdk.Msg = &MsgReportMissed{}
	_ sdk.Msg = &MsgReportInvalid{}
)

// MsgRegisterSidecars defines a message to register relayer and validator public keys.
type MsgRegisterSidecars struct {
	ValAddr         sdk.ValAddress `json:"val_addr"`
	RelayerPubKey   []byte         `json:"relayer_pub_key"`
	ValidatorPubKey []byte         `json:"validator_pub_key"`
}

func NewMsgRegisterSidecars(valAddr sdk.ValAddress, relayerPubKey, validatorPubKey []byte) *MsgRegisterSidecars {
	return &MsgRegisterSidecars{
		ValAddr:         valAddr,
		RelayerPubKey:   relayerPubKey,
		ValidatorPubKey: validatorPubKey,
	}
}

// Route returns the message route.
func (msg MsgRegisterSidecars) Route() string { return RouterKey }

// Type returns the message type.
func (msg MsgRegisterSidecars) Type() string { return "register_sidecars" }

// ValidateBasic performs basic validation of the message.
func (msg MsgRegisterSidecars) ValidateBasic() error {
	if msg.ValAddr.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "validator address cannot be empty")
	}
	return nil
}

// GetSignBytes returns the canonical byte representation of the message.
func (msg MsgRegisterSidecars) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the signers of the message.
func (msg MsgRegisterSidecars) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(msg.ValAddr)}
}

// MsgHeartbeat defines a message to send a heartbeat.
type MsgHeartbeat struct {
	ValAddr         sdk.ValAddress `json:"val_addr"`
	OriginHeightsJSON string         `json:"origin_heights_json"`
	Sig             []byte         `json:"sig"`
}

func NewMsgHeartbeat(valAddr sdk.ValAddress, originHeightsJSON string, sig []byte) *MsgHeartbeat {
	return &MsgHeartbeat{
		ValAddr:         valAddr,
		OriginHeightsJSON: originHeightsJSON,
		Sig:             sig,
	}
}

// Route returns the message route.
func (msg MsgHeartbeat) Route() string { return RouterKey }

// Type returns the message type.
func (msg MsgHeartbeat) Type() string { return "heartbeat" }

// ValidateBasic performs basic validation of the message.
func (msg MsgHeartbeat) ValidateBasic() error {
	if msg.ValAddr.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "validator address cannot be empty")
	}
	return nil
}

// GetSignBytes returns the canonical byte representation of the message.
func (msg MsgHeartbeat) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the signers of the message.
func (msg MsgHeartbeat) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(msg.ValAddr)}
}

// MsgReportMissed defines a message to report a missed duty.
type MsgReportMissed struct {
	Route                 Route          `json:"route"`
	MsgID                 uint64         `json:"msg_id"`
	AssignedVal           sdk.ValAddress `json:"assigned_val"`
	OriginProof           []byte         `json:"origin_proof"`
	DestNonInclusionProof []byte         `json:"dest_non_inclusion_proof"`
	Signer                sdk.AccAddress `json:"signer"`
}

func NewMsgReportMissed(route Route, msgID uint64, assignedVal sdk.ValAddress, originProof, destNonInclusionProof []byte, signer sdk.AccAddress) *MsgReportMissed {
	return &MsgReportMissed{
		Route:                 route,
		MsgID:                 msgID,
		AssignedVal:           assignedVal,
		OriginProof:           originProof,
		DestNonInclusionProof: destNonInclusionProof,
		Signer:                signer,
	}
}

// Route returns the message route.
func (msg MsgReportMissed) Route() string { return RouterKey }

// Type returns the message type.
func (msg MsgReportMissed) Type() string { return "report_missed" }

// ValidateBasic performs basic validation of the message.
func (msg MsgReportMissed) ValidateBasic() error {
	if msg.AssignedVal.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "assigned validator address cannot be empty")
	}
	if msg.Signer.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "signer address cannot be empty")
	}
	return nil
}

// GetSignBytes returns the canonical byte representation of the message.
func (msg MsgReportMissed) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the signers of the message.
func (msg MsgReportMissed) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

// MsgReportInvalid defines a message to report an invalid relay.
type MsgReportInvalid struct {
	Route            Route          `json:"route"`
	MsgID            uint64         `json:"msg_id"`
	AssignedVal      sdk.ValAddress `json:"assigned_val"`
	DestFailureProof []byte         `json:"dest_failure_proof"`
	Signer           sdk.AccAddress `json:"signer"`
}

func NewMsgReportInvalid(route Route, msgID uint64, assignedVal sdk.ValAddress, destFailureProof []byte, signer sdk.AccAddress) *MsgReportInvalid {
	return &MsgReportInvalid{
		Route:            route,
		MsgID:            msgID,
		AssignedVal:      assignedVal,
		DestFailureProof: destFailureProof,
		Signer:           signer,
	}
}

// Route returns the message route.
func (msg MsgReportInvalid) Route() string { return RouterKey }

// Type returns the message type.
func (msg MsgReportInvalid) Type() string { return "report_invalid" }

// ValidateBasic performs basic validation of the message.
func (msg MsgReportInvalid) ValidateBasic() error {
	if msg.AssignedVal.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "assigned validator address cannot be empty")
	}
	if msg.Signer.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "signer address cannot be empty")
	}
	return nil
}

// GetSignBytes returns the canonical byte representation of the message.
func (msg MsgReportInvalid) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the signers of the message.
func (msg MsgReportInvalid) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
