package handler

// FileCryptTimeout exposes fileCryptTimeout so the deadline tests can assert
// against the configured budget instead of a hard-coded one. Intended for use
// in tests only.
const FileCryptTimeout = fileCryptTimeout
