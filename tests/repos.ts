import * as pulumi from '@pulumi/pulumi'

pulumi.runtime.setMocks({
    newResource: function (type: string, name: string, inputs: any): { id: string, state: any } {
        return {
            id: inputs.name + "_id",
            state: inputs,
        }
    },
    call: function (token: string, args: any, provider?: string) {
        return args;
    }
})

import * as repos from '../index'

describe('Repositories', function () {
    const repositories = repos.repositories
        repositories.map(repo => {
            pulumi.all([repo.name, repo.visibility]).apply(([name, visibility]) => {
                it(`${name} must be public`, function (done) {
                if (visibility == 'public') {
                    done()
                } else {
                    done(new Error(`visibilty should be public for ${name}`))
                }
            })
        })
    })
})
