# Dendrite - Research Project Modifications

This repository is a modified fork of the original Matrix Dendrite homeserver, which can be found [here](https://github.com/element-hq/dendrite). The code has been modified to serve as a component in the "Achieving User Pseudonymity in Federated End-to-End Encrypted Messaging Platforms" research project. For an overview of the project, see the [main project README](https://github.com/Ijnek117/anonymous-matrix).

---

## 1. Introduction

Dendrite was chosen for this project because [MSC4014: Pseudonymous Identities](https://github.com/matrix-org/matrix-spec-proposals/pull/4014), a Matrix Proposal that breaks the association between a user's ID and their activity in a room, had already been implemented in it.

Part of the changes made by MSC4014 was the modification of the `sender` field from a Matrix User ID to a `sender_key` (an ed25519 public key) scoped to a specific per-room, per-user identity. This means that events, such as messages, were no longer directly associated with a user's User ID (their persistent Matrix identity). However, events of type `m.room.member` still contained a mapping between the User ID and the new sender IDs to allow for routing, cross-room tracking, and other purposes.
This research project extends the changes made in MSC4014 by removing all plaintext Matrix User IDs, including those in the User ID to sender ID mappings.

## 2. Summary of Changes

The following changes were made to this repository for the research project:

* Added a new file in `clientapi/routing/` to handle a custom API endpoint for fetching a remote server's TLS certificates and removed display names from join events.
* Modified `federationapi/api` to handle encrypted User ID invitations.
* Modified `federationapi/internal` by adding `SendEncryptedInvite`, which takes an invitee of type `spec.EncryptedUserID`, and added support for join events based on Sender IDs.
* Modified the `/invite/{roomID}/{userID}` endpoint in `federationapi/routing` to handle encrypted User IDs.
* Modified the federation `make_join` endpoint in `federationapi/routing/routing.go` to use a Sender ID instead of a User ID in the path.
* Modified `roomserver/api` to handle encrypted User ID invitations.
* Added new request and input types to `roomserver/api/perform.go` to support encrypted invites.
* Modified `roomserver/internal` by adding the `PerformEncryptedInvite` implementation, storing a placeholder for remote User IDs.
* Added a new configuration file `research.yaml` for development.

## 3. Rationale for Modifications

These changes were necessary to prevent remote homeservers from tracing user activity across rooms. Join events are now modified to only send the homeserver domain in the `mxid_mapping`, rather than the full User ID. Consequently, a remote server only knows that **some** user is joining from homeserver A, but not which specific user it is.

Support for join events with client-encrypted User IDs has also been added. This includes the necessary User ID decryption on the receiving server and a new mechanism for fetching a remote server's TLS certificate, which is required for the encryption.

For federation purposes, Sender IDs need to be stored along with the server name from which they originate. This is currently achieved using the existing infrastructure by creating placeholder User IDs from the incomplete ones received in the `mxid_mapping` of `m.room.member` events. This approach is temporary and is expected to be replaced by a larger refactoring effort that will remove the use of the `spec.UserID` type from federation methods and related database schemas.

## 4. Key Files Modified

This table provides a more detailed reference to the files and code that were changed or added.

| File Path | Change Description |
| :--- | :--- |
| `clientapi/routing/joinroom.go` | Removed the `displayname` from being added to the content of join events. |
| `clientapi/routing/membership.go` | Modified the `SendInvite` function to support encrypted User ID invitations for pseudonymous rooms by adding a new `sendEncryptedInvite` helper and checking the room version. Also removed the `displayname` from direct membership events. |
| `clientapi/routing/routing.go` | Added a new API route `/server_tls_keys/{serverName}` to handle requests for remote server TLS certificates, calling the new `QueryServerTLSCertificate` handler. |
| `clientapi/routing/server_tls_key.go` | **(New File)** Implemented the `QueryServerTLSCertificate` handler for the new client API endpoint. It defines the `serverTLSCertResponse` struct and contains logic to call the federation API and format the certificate details into a JSON response. |
| `federationapi/api/api.go` | Added the `SendEncryptedInvite` method to the `RoomserverFederationAPI` interface to support sending invites with encrypted User IDs over federation. |
| `federationapi/internal/perform.go` | Implemented the `SendEncryptedInvite` method on the `FederationInternalAPI` struct, which handles the logic for sending an invite with an encrypted User ID to a remote homeserver. |
| `federationapi/routing/join.go` | Modified the `/send_join` handler to call `gomatrixserverlib.HandlePseudoSendJoin` for rooms with pseudonymous IDs. |
| `federationapi/routing/routing.go` | Modified the `/invite/{roomID}/{userID}` endpoint to decrypt encrypted User IDs using the server's private key. Replaced the `/make_join/{roomID}/{userID}` endpoint with `/make_join/{roomID}/{senderID}` to use Sender IDs instead of User IDs in the path for join requests. |
| `research.yaml` | **(New File)** Added a new configuration file for research project development. |
| `roomserver/api/api.go` | Added the `PerformEncryptedInvite` method to the `ClientRoomserverAPI` interface to expose the new encrypted invite functionality. |
| `roomserver/api/perform.go` | Added new request and input types (`PerformEncryptedInviteRequest`, `EncryptedInviteInput`) to support performing invites with encrypted User IDs. |
| `roomserver/internal/api.go` | Implemented the `PerformEncryptedInvite` method on the `RoomserverInternalAPI` struct to handle the new encrypted invite requests. |
| `roomserver/internal/input/input_events.go` | Modified event processing to handle `mxid_mapping` from remote servers differently. It now stores a placeholder User ID (`@a:<domain>`) for remote users instead of their full Matrix ID to prevent cross-room tracking. See Rationale. |
| `roomserver/internal/perform/perform_create_room.go` | Modified room creation to remove the `DisplayName` from the creator's initial membership event and to only include the homeserver domain (not the full User ID) in the `mxid_mapping`. |
| `roomserver/internal/perform/perform_invite.go` | Added the `PerformEncryptedInvite` implementation, which contains the core logic for creating and processing an invitation with an encrypted User ID within the roomserver. |
| `roomserver/internal/perform/perform_join.go` | Modified the join logic to only include the joining user's homeserver domain (not their full User ID) in the `mxid_mapping` of their membership event. |
| `userapi/consumers/roomserver.go` | Added a nil-check for the sender's User ID before evaluating push rules to prevent panics. |