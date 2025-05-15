package zfile

// File size unit constants for use in size calculations and formatting.
const (
	// BYTE represents 1 byte
	BYTE = 1 << (iota * 10)
	// KB represents 1 kilobyte (1024 bytes)
	KB
	// MB represents 1 megabyte (1024 kilobytes)
	MB
	// GB represents 1 gigabyte (1024 megabytes)
	GB
	// TB represents 1 terabyte (1024 gigabytes)
	TB
)
