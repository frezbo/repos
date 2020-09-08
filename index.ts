import * as github from '@pulumi/github'

const ghProvider = new github.Provider('github', { token: process.env.GITHUB_TOKEN, owner: 'frezbo' })

const repos = [
  'resume',
  'repos',
  'rpminfo'
]

const repositories: github.Repository[] = []

repos.map(repo => {
  repositories.push(new github.Repository(repo, {
    visibility: 'public',
    name: repo
  }, { provider: ghProvider, import: repo }))
})

export { repositories }
