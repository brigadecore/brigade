package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/deis/acid/pkg/webhook"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "required arg: Git SHA")
		os.Exit(1)
	}
	commit := os.Args[1]

	data, err := ioutil.ReadFile("./_functional_tests/test-repo.json")
	if err != nil {
		panic(err)
	}

	pushHook := &webhook.PushHook{}
	if err := json.Unmarshal(data, pushHook); err != nil {
		panic(err)
	}

	// Set the commit ID:
	pushHook.HeadCommit.Id = commit

	out, err := json.MarshalIndent(pushHook, "", "  ")
	if err != nil {
		panic(err)
	}

	secret := getSecret(pushHook)
	hmac := webhook.SHA1HMAC([]byte(secret), out)

	ioutil.WriteFile("./_functional_tests/test-repo-generated.json", out, 0755)
	ioutil.WriteFile("./_functional_tests/test-repo-generated.hash", []byte(hmac), 0755)

	fmt.Fprintln(os.Stdout, string(out))
	fmt.Fprintln(os.Stdout, hmac)
}

func getSecret(pushHook *webhook.PushHook) string {
	pname := "acid-" + webhook.ShortSHA(pushHook.Repository.FullName)
	proj, err := webhook.LoadProjectConfig(pname, "default")
	if err != nil {
		panic(err)
	}
	return proj.SharedSecret
}
