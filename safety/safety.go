package safety

import (
  "log"
  "runtime/debug"
)

// GoSafe runs a function in a goroutine and recovers from any panic that occurs
// Usage example:
// safety.GoSafe(func() {
//   // Your code that might panic
// })
func GoSafe(fn func()) {
  go func() {
    defer RecoverPanic()
    fn()
  }()
}

// RecoverPanic recovers from a panic and logs the error
// Usage example:
// go func() {
//   defer safety.RecoverPanic()
//   // Your code that might panic
// }()
func RecoverPanic() {
  if r := recover(); r != nil {
    log.Printf("Recovered from panic: %v\n%s", r, debug.Stack())
  }
}

// RecoverPanicWithHandler recovers from a panic and calls the provided handler function
// Usage example:
// go func() {
//   defer safety.RecoverPanicWithHandler(func(r interface{}, stack []byte) {
//     // Handle the panic here
//     log.Printf("Recovered from panic: %v\n%s", r, stack)
//   })
//   // Your code that might panic
// }()
func RecoverPanicWithHandler(handler func(interface{}, []byte)) {
  if r := recover(); r != nil {
    stack := debug.Stack()
    handler(r, stack)
  }
}
