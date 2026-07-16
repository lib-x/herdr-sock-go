// Package herdrsock provides a Go client for Herdr's local socket API.
//
// Herdr speaks newline-delimited JSON on its API socket. Most calls open one
// connection, write one request line, and read one response line. Event
// subscriptions keep the connection open after the subscription ack.
//
// CurrentHerdrVersion identifies the Herdr source schema used by the typed API.
// Herdr 0.7.3 and 0.7.4 both use CurrentProtocol 16; callers that depend on
// 0.7.4-only methods should also inspect Ping().Version.
package herdrsock
