// Package streamwork provides a convenient way to use streaming workers:
// Concurrent functions passing data down the stream to accomplish a job together.
// It is particularly well suited for jobs with multiple slow I/O dependencies.
// For example: reading rows from a CSV, transforming each one through an API call, then saving each result in database.
// Read the Example to see how to use it.
package streamwork
