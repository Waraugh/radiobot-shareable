# RADIO BOB! 2000er Rock shareable player

A small installable web player for RADIO BOB! 2000er Rock.

## Deploy to Netlify

This package includes a Go Netlify Function, so do not use Netlify Drop for this version.

1. Put the contents of this folder in a GitHub/GitLab/Bitbucket repository.
2. In Netlify choose **Add new project → Import an existing project**.
3. Select the repository and publish it. `netlify.toml` already defines the publish and functions directories.
4. Open the generated Netlify URL. The player should show live ICY `StreamTitle` metadata and play the RADIO BOB stream.
5. Share that URL. On iPhone, Share → Add to Home Screen installs it as a standalone PWA.

No environment variables or API keys are required.
