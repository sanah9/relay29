package khatru29

import (
	"context"

	"github.com/fiatjaf/khatru"
	"github.com/fiatjaf/relay29"
	"github.com/nbd-wtf/go-nostr"
)

// relayAdapter wraps *khatru.Relay and adapts its methods to the interface expected by relay29.State.
type relayAdapter struct {
	*khatru.Relay
}

// BroadcastEvent satisfies the relay29.State expected signature (no return value).
func (r *relayAdapter) BroadcastEvent(evt *nostr.Event) {
	_ = r.Relay.BroadcastEvent(evt)
}

// AddEvent simply proxies to the embedded khatru.Relay implementation.
func (r *relayAdapter) AddEvent(ctx context.Context, evt *nostr.Event) (bool, error) {
	return r.Relay.AddEvent(ctx, evt)
}

func Init(opts relay29.Options) (*khatru.Relay, *relay29.State) {
	pubkey, _ := nostr.GetPublicKey(opts.SecretKey)

	// create a new relay29.State
	state := relay29.New(opts)

	// create a new khatru relay
	relay := khatru.NewRelay()
	relay.Info.PubKey = pubkey
	relay.Info.SupportedNIPs = append(relay.Info.SupportedNIPs, 29)

	// assign khatru relay to relay29.State
	state.Relay = &relayAdapter{Relay: relay}

	// provide GetAuthed function
	state.GetAuthed = khatru.GetAuthed

	// apply basic relay policies
	relay.StoreEvent = append(relay.StoreEvent, state.DB.SaveEvent)
	relay.QueryEvents = append(relay.QueryEvents,
		state.NormalEventQuery,
		state.MetadataQueryHandler,
		state.AdminsQueryHandler,
		state.MembersQueryHandler,
		state.RolesQueryHandler,
	)
	relay.DeleteEvent = append(relay.DeleteEvent, state.DB.DeleteEvent)
	relay.RejectFilter = append(relay.RejectFilter,
		state.RequireKindAndSingleGroupIDOrSpecificEventReference,
	)
	relay.RejectEvent = append(relay.RejectEvent,
		state.RequireHTagForExistingGroup,
		state.RequireModerationEventsToBeRecent,
		state.RestrictWritesBasedOnGroupRules,
		state.RestrictInvalidModerationActions,
		state.PreventWritingOfEventsJustDeleted,
		state.CheckPreviousTag,
	)
	relay.OnEventSaved = append(relay.OnEventSaved,
		state.ApplyModerationAction,
		state.ReactToJoinRequest,
		state.ReactToLeaveRequest,
		state.AddToPreviousChecking,
	)
	relay.OnConnect = append(relay.OnConnect, khatru.RequestAuth)

	return relay, state
}
