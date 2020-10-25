import * as github from '@pulumi/github'

const ghProvider = new github.Provider('github', { token: process.env.GITHUB_TOKEN, owner: 'frezbo' })

const repos = [
  { name: 'resume', description: 'Repository to hold personal Resume' },
  { name: 'repos', description: 'Manage personal repositories' },
  { name: 'rpminfo', description: 'Retrieve RPM packages list from yum repo' },
  { name: 'infra-dns', description: 'Project to manage personal DNS' }
]

const repositories: github.Repository[] = []

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
})

export { repositories }

