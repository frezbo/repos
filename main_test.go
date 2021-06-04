package main

import (
	"os"
	"sync"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
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

func (mocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	return args.Name + "_id", args.Inputs, nil
}

func (mocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	return resource.PropertyMap{}, nil
}

func TestCreateRepositories(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		repos, err := createRepositories(ctx)
		assert.NoError(t, err)

		var wg sync.WaitGroup
		wg.Add(len(repos))

		for _, repo := range repos {
			pulumi.All(repo.DeleteBranchOnMerge).ApplyT(func(all []interface{}) error {
				deleteBranchOnMerge := all[0].(*bool)
				assert.Equalf(t, deleteBranchOnMerge, &[]bool{true}[0], "Expected delete on merge to be: %s", "true")
				wg.Done()
				return nil
			})
		}

		wg.Wait()
		return nil
	}, pulumi.WithMocks("repos", "somestack", mocks(0)))
	assert.NoError(t, err)
}
