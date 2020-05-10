This example will demo how the `smcache` library interacts with GCP's secret manager. It can be useful
to help understand how this code behaves with GCP Secret Manager.

Before running this on your own GCP server, you'll want to edit the
`projectID := "example-project-1234"`, and replace the example-project-1234 with your GCP project name.

If you compile and run this on a GCP host that has access to interact with the Secret Manager, it will:

1. Create a new Secret called `testsite-www_example_com`, and store a hard-coded string into it.
2. Read the secret, and print it to stdout.
3. Write different data to the same key, which should create a `version 2` of the secret.
4. Read the secret again, and print the new string to stdout.

It will not delete the secret (the code to do so is present, but comments out).
