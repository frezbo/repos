package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/pulumi/pulumi-github/sdk/v2/go/github"
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"
)

type repository struct {
	Name        string
	Description string
	Secrets     []string
}

const (
	defaultBranch         = "main"
	repoVisibility        = "public"
	pulumiTokenEnvVarName = "PULUMI_ACCESS_TOKEN"
)

var repositories = []repository{
	{
		Name:        "resume",
		Description: "Repository to hold personal Resume",
	},
	{
		Name:        "repos",
		Description: "Manage personal repositories",
	},
	{
		Name:        "rpminfo",
		Description: "Retrieve RPM packages list from yum repo",
	},
	{
		Name:        "infra-dns",
		Description: "Project to manage personal DNS",
		Secrets: []string{
			"CLOUDFLARE_API_TOKEN",
			"PULUMI_ACCESS_TOKEN",
		},
	},
	{
		Name:        "infra-do",
		Description: "Project to manage DigitalOcean Resources",
		Secrets: []string{
			"DIGITALOCEAN_TOKEN",
			"PULUMI_ACCESS_TOKEN",
		},
	},
	{
		Name:        "docker-actions-test",
		Description: "Project to test Multi-Arch docker builds and push to GHCR",
		Secrets: []string{
			"GHCR_ACCESS_TOKEN",
		},
	},
	{
		Name:        "openfaas-template-static",
		Description: "OpenFaaS template to serve static files",
	},
	{
		Name:        "ansible-workstation",
		Description: "Manage workstation configuration",
	},
	{
		Name:        "rss-feeds",
		Description: "RSS feed manager for mattermost RSS plugin",
	},
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		repos, err := createRepositories(ctx)
		if err != nil {
			return err
		}
		var sshCloneURLs = pulumi.StringArray{}
		for _, repo := range repos {
			sshCloneURLs = append(sshCloneURLs, repo.SshCloneUrl)
		}
		ctx.Export("sshCloneURLs", sshCloneURLs)
		return nil
	})
}

func createRepositories(ctx *pulumi.Context) ([]*github.Repository, error) {
	githubAccessToken := os.Getenv("PERSONAL_ACCESS_TOKEN")
	if githubAccessToken == "" {
		return nil, errors.New("PERSONAL_ACCESS_TOKEN environment variable not set")
	}
	provider, err := github.NewProvider(ctx, "github", &github.ProviderArgs{
		Owner: pulumi.String("frezbo"),
		Token: pulumi.String(githubAccessToken),
	})
	if err != nil {
		return nil, err
	}
	outputs := []*github.Repository{}
	for _, repository := range repositories {
		repo, err := github.NewRepository(ctx, repository.Name, &github.RepositoryArgs{
			AllowMergeCommit:    pulumi.Bool(true),
			AllowRebaseMerge:    pulumi.Bool(true),
			AllowSquashMerge:    pulumi.Bool(true),
			Archived:            pulumi.Bool(false),
			AutoInit:            pulumi.Bool(false),
			DefaultBranch:       pulumi.String(defaultBranch),
			DeleteBranchOnMerge: pulumi.Bool(true),
			Description:         pulumi.String(repository.Description),
			Name:                pulumi.String(repository.Name),
			Visibility:          pulumi.String(repoVisibility),
			HasDownloads:        pulumi.Bool(true),
			HasIssues:           pulumi.Bool(true),
			HasProjects:         pulumi.Bool(true),
			HasWiki:             pulumi.Bool(true),
			IsTemplate:          pulumi.Bool(false),
			VulnerabilityAlerts: pulumi.BoolPtr(true),
		}, pulumi.Provider(provider))
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, repo)
		if repository.Secrets != nil {
			for _, secretEnv := range repository.Secrets {
				secretName := fmt.Sprintf("%s-%s", repository.Name, secretEnv)
				_, err := github.NewActionsSecret(ctx, secretName, &github.ActionsSecretArgs{
					PlaintextValue: pulumi.String(secretEnvFromRepo(repository.Name, secretEnv)),
					Repository:     pulumi.String(repository.Name),
					SecretName:     pulumi.String(secretEnv),
				}, pulumi.Provider(provider))
				if err != nil {
					return nil, err
				}
			}
		}
		_, err = github.NewBranchProtection(ctx, repository.Name, &github.BranchProtectionArgs{
			EnforceAdmins:        pulumi.Bool(true),
			Pattern:              pulumi.String(defaultBranch),
			RepositoryId:         repo.NodeId,
			RequireSignedCommits: pulumi.Bool(true),
			RequiredStatusChecks: github.BranchProtectionRequiredStatusCheckArray{github.BranchProtectionRequiredStatusCheckArgs{
				Strict: pulumi.Bool(true),
			}},
		}, pulumi.Provider(provider))
		if err != nil {
			return nil, err
		}
	}
	return outputs, nil
}

func secretEnvFromRepo(repo string, secret string) string {
	if secret == pulumiTokenEnvVarName {
		return os.Getenv(pulumiTokenEnvVarName)
	}
	var repoNameUpper = strings.ToUpper(strings.ReplaceAll(repo, "-", "_"))
	secretFromEnv := os.Getenv(fmt.Sprintf("%s_%s", repoNameUpper, secret))
	return secretFromEnv
}
