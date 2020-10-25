import * as github from '@pulumi/github'
import { env } from 'process'

const ghProvider = new github.Provider('github', { token: process.env.GITHUB_TOKEN, owner: 'frezbo' })

const repos = [
  { name: 'resume', description: 'Repository to hold personal Resume', secrets: [] },
  { name: 'repos', description: 'Manage personal repositories', secrets: [] },
  { name: 'rpminfo', description: 'Retrieve RPM packages list from yum repo', secrets: [] },
  { name: 'infra-dns', description: 'Project to manage personal DNS', secrets: [ 'CLOUDFLARE_API_TOKEN', 'PULUMI_ACCESS_TOKEN' ] }
]

const repositories: github.Repository[] = []
const secrets: github.ActionsSecret[] = []

repos.map(repo => {
  repositories.push(new github.Repository(repo.name, {
    visibility: 'public',
    name: repo.name,
    allowMergeCommit: true,
    allowRebaseMerge: true,
    allowSquashMerge: true,
    archived: false,
    autoInit: false,
    defaultBranch: 'main',
    deleteBranchOnMerge: true,
    description: repo.description,
    hasDownloads: true,
    hasIssues: true,
    hasProjects: true,
    hasWiki: true,
    isTemplate: false
  }, { provider: ghProvider }))
  repo.secrets.map(secret => {
    let envValue
    if (secret == 'PULUMI_ACCESS_TOKEN') {
      envValue = process.env.secret
    } else {
      envValue = process.env[`${repo.name.replace('-', '_').toUpperCase()}_${secret}`]
    }
    if (envValue === undefined) {
      envValue = ""
    }
    secrets.push(new github.ActionsSecret(secret, {
      plaintextValue: envValue,
      repository: repo.name,
      secretName: secret
    }))
  })
})

export { repositories }
