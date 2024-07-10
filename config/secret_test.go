package config

import (
	"context"
	"fmt"
)

func ExampleResolveSecretFromGCP() {
	ctx := context.Background()
	fmt.Println(ResolveSecretFromGCP(ctx, "not-a-gcp-secret")) // the value does not start with the expected prefix, it will be printed as-is

	// this would fetch the secret from GCP, if we had some working credentials
	// fmt.Println(
	// 	ResolveSecretFromGCP(
	// 		ctx,
	// 		"gcp-secret://projects/my-project/secrets/my-secret/versions/latest",
	// 	),
	// )

	// Output:
	// not-a-gcp-secret <nil>
}
