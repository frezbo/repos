import * as github from "@pulumi/github";

const provider = new github.Provider('github', { token: process.env.GITHUB_TOKEN, owner: 'frezbo' })

const repos = [
    'resume',
    'repos',
    'rpminfo'
]

const repositories = repos.map(repo => {
    new github.Repository(repo, {
        visibility: 'public',
        name: repo
    })
})

export { repositories }
