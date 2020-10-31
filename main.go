package main

import (
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
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		outputs := pulumi.StringArray{}
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
			})
			if err != nil {
				return err
			}
			outputs = append(outputs, repo.SshCloneUrl)
			if repository.Secrets != nil {
				for _, secretEnv := range repository.Secrets {
					_, err := github.NewActionsSecret(ctx, secretEnv, &github.ActionsSecretArgs{
						PlaintextValue: pulumi.String(secretEnvFromRepo(repository.Name, secretEnv)),
						Repository:     pulumi.String(repository.Name),
						SecretName:     pulumi.String(secretEnv),
					})
					if err != nil {
						return err
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
			})
			if err != nil {
				return err
			}
		}

		ctx.Export("sshCloneURLs", outputs)
		return nil
	})
}

func secretEnvFromRepo(repo string, secret string) string {
	var repoNameUpper = strings.ToUpper(strings.ReplaceAll(repo, "-", "_"))
	secretFromEnv := os.Getenv(fmt.Sprintf("%s_%s", repoNameUpper, secret))
	return secretFromEnv
}
