package main

import (
	"os"
	"sync"
	"testing"

	"github.com/pulumi/pulumi/sdk/v2/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"
	"github.com/stretchr/testify/assert"
)

func TestSecretEnvFromRepo(t *testing.T) {
	t.Cleanup(testEnvCleanup)
	os.Setenv("SOME_REPO_SOME_SECRET", "some-repo-secret")
	os.Setenv(pulumiTokenEnvVarName, "pulumi-secret")
	cases := map[string]struct {
		Repo   string
		Secret string
		Value  string
	}{
		"CanRetrieveEnvVarForARepoWithEnvVarSet": {
			Repo:   "some-repo",
			Secret: "SOME_SECRET",
			Value:  "some-repo-secret",
		},
		"ReturnsEmptyStringForEnvVarNotSet": {
			Repo:   "some-new-repo",
			Secret: "SOME_SECRET",
			Value:  "",
		},
		"ReturnsPulumiTokenWhenPulumiEnvVarIsReferenced": {
			Repo:   "some-repo",
			Secret: pulumiTokenEnvVarName,
			Value:  "pulumi-secret",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if envVar := secretEnvFromRepo(tc.Repo, tc.Secret); envVar != tc.Value {
				t.Errorf("Expected: %s, got: %s", tc.Value, envVar)
			}
		})
	}
}

func testEnvCleanup() {
	os.Unsetenv("SOME_REPO_SOME_SECRET")
	os.Unsetenv(pulumiTokenEnvVarName)
}

type mocks int

func (mocks) NewResource(typeToken, name string, inputs resource.PropertyMap, provider, id string) (string, resource.PropertyMap, error) {
	return name + "_id", inputs, nil
}

func (mocks) Call(token string, args resource.PropertyMap, provider string) (resource.PropertyMap, error) {
	return args, nil
}

func TestCreateRepositories(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		repos, err := createRepositories(ctx)
		assert.NoError(t, err)

		var wg sync.WaitGroup
		wg.Add(len(repos))

		for _, repo := range repos {
			pulumi.All(repo.DefaultBranch).ApplyT(func(all []interface{}) error {
				defaultBranch := all[0].(string)
				assert.Equalf(t, "main", defaultBranch, "Expected default branch to be: %s", "main")
				wg.Done()
				return nil
			})
		}

		wg.Wait()
		return nil
	}, pulumi.WithMocks("repos", "somestack", mocks(0)))
	assert.NoError(t, err)
}
