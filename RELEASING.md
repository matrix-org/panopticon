# Releasing (and Deploying) Panopticon — some brief notes

1. Ensure that you are on a recently-updated copy of the `master` branch and that CI is happy.
2. Based on the changes since the last release, add an entry at the top of the `CHANGELOG.md`, following the style of the prior entries.
3. Set a variable for the version you are releasing, for convenience: `ver=x.y.z`.
4. `git add CHANGELOG.md && git commit -m $ver && git push`
5. Give [the changelog](https://github.com/matrix-org/panopticon/blob/master/CHANGELOG.md) a quick check to ensure everything is in order.
6. When you're ready, tag the release with `git tag -s v$ver` and push it with `git push origin tag v$ver`.
7. Create a release on GitHub: `xdg-open https://github.com/matrix-org/panopticon/releases/edit/v$ver`
8. With any luck, *GitHub Actions* will spring into action, building the Docker images and pushing them to [Docker Hub](https://hub.docker.com/r/matrixdotorg/panopticon/tags?page=1&ordering=last_updated).


[Private, infrastructure-specific, instructions](https://gitlab.matrix.org/new-vector/internal/-/wikis/Panopticon) are available to internal members for deploying the 'official' deployment of Panopticon.
