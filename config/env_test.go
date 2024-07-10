package config

import (
	"fmt"
	"os"
)

func ExampleGetEnvOrDefault() {
	os.Unsetenv("FOO")
	os.Setenv("BAR", "bar")

	fmt.Println(GetEnvOrDefault("FOO", "foo"))    // the env var does not exist, it will print the default value
	fmt.Println(GetEnvOrDefault("BAR", "notbar")) // the env var does exist, it will print the env value
	// Output:
	// foo
	// bar
}

func ExampleGetEnvOrError() {
	os.Unsetenv("FOO")
	os.Setenv("BAR", "bar")

	fmt.Println(GetEnvOrError("FOO")) // the env var does not exist, it will print the error
	fmt.Println(GetEnvOrError("BAR")) // the env var does exist, it will print the env value
	// Output:
	// missing env var "FOO"
	// bar <nil>
}
