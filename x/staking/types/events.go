package types

// staking module event types
const (
	EventTypeCompleteUnbonding    = "complete_unbonding"
	EventTypeCompleteRedelegation = "complete_redelegation"
	EventTypeCreateValidator      = "create_validator"
	EventTypeEditValidator        = "edit_validator"
	EventTypeDelegate             = "delegate"
	EventTypeUnbond               = "unbond"
	EventTypeForceUnbond          = "force_unbond"
	EventTypeRedelegate           = "redelegate"
	EventTypeScheduleUnbond       = "schedule_unbond"
	EventTypeUnscheduleUnbond     = "unschedule_unbond"

	AttributeKeyValidator         = "validator"
	AttributeKeyCommissionRate    = "commission_rate"
	AttributeKeyMinSelfDelegation = "min_self_delegation"
	AttributeKeySrcValidator      = "source_validator"
	AttributeKeyDstValidator      = "destination_validator"
	AttributeKeyDelegator         = "delegator"
	AttributeKeyCompletionTime    = "completion_time"
	AttributeValueCategory        = ModuleName
)
