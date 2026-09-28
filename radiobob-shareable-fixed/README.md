# RADIO BOB Shareable Player

Repository root must contain:

- `netlify.toml`
- `go.mod`
- `site/index.html`
- `site/manifest.webmanifest`
- `site/sw.js`
- `site/icon-180.png`
- `site/icon-192.png`
- `site/icon-512.png`
- `netlify/functions/nowplaying/main.go`

Netlify configuration:
- Publish directory: `site`
- Functions directory: `netlify/functions`
- Function endpoint: `/.netlify/functions/nowplaying`
- Friendly endpoint: `/api/now-playing`

Push these files to the root of the GitHub repository. Netlify should automatically redeploy.
