package chatroomsummary

import "embed"

// FS exposes shared robot templates to packages outside this directory.
//
//go:embed chat_room_summary.html
var FS embed.FS
