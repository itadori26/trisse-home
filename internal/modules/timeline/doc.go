// Package timeline is the event timeline module (#74): a chronological log
// of notable events (services offline/online, containers, backups and
// heartbeat checks, certificates, disks, security alerts). It receives the
// same kernel observations and alerts as the notifications module; other
// modules do not know it exists.
//
// Ring: module wiring.
package timeline
