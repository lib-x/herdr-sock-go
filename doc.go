// Package herdrsock provides a Go client for Herdr's local socket API.
//
// Herdr speaks newline-delimited JSON on its API socket. Most calls open one
// connection, write one request line, and read one response line. Event
// subscriptions keep the connection open after the subscription ack.
//
// Use RequireCurrentProtocol before depending on behavior tied to this package's
// generated protocol version.
package herdrsock
