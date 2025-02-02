# Setting up tailwind

_One option is to use the CDN, but not suitable for production. Instead, what we do is:_

## Setting up TailwindCSS4 using Standalone CLI (no Node/npm required)

See: https://tailwindcss.com/blog/standalone-cli

Step 1: Install the standalone CLI by checking the latest release on Github, make it executable and add it to the project:

```
curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-arm64
chmod +x tailwindcss-macos-arm64
mv tailwindcss-macos-arm64 tailwindcss
```

**Side note here**: there is no more `./tailwindcss init` command and the `tailwind.config.js` file is no longer needed.
The binary will use sensible paths to scan for html files. However, the `tailwind.config.js` file can always be created manually if needed (e.g. to specify a content that includes non-'\*html' templates, like templ)

Step 2: And now we can start a watcher to keep our `./public/styles.css` updated:

```
./tailwindcss -i ./web/views/css/styles.css -o ./public/styles.css --watch
```

Step 3: For production, compile and minify the CSS file:

```
./tailwindcss -i ./web/views/css/styles.css -o ./public/styles.css --minify
```

## Alternative: Setting up TailwindCSS4 using tailwind CLI with Node.js (not used here)

First, we'll need to

```
npm init -y
```

and then

```
npm install tailwindcss @tailwindcss/cli
```

to install the cli that we can run with npx.

Then add the following to `./web/views/css/styles.css`:

```
@import "tailwindcss";
```

Next, run:

```
npx @tailwindcss/cli -i ./web/views/css/styles.css -o ./public/styles.css --watch
```

Add `<link href="./public/styles.css" rel="stylesheet">` to the `<head>`.

Source: https://tailwindcss.com/docs/installation/tailwind-cli

## Alternative: Setting up Tailwind 3 (not used here)

```
npm install -D tailwindcss@3
npx tailwindcss init
```

Add `./web/views/*.html` to tailwind.config.js under `content`.

Then add the following to `./web/views/css/styles.css`:

```
@tailwind base;
@tailwind components;
@tailwind utilities;
```

Run the CLI to scan the template files and build the CSS:

```
npx tailwindcss -i ./web/views/css/styles.css -o ./public/styles.css --watch
```

And add `<link href="./public/styles.css" rel="stylesheet">` to the `<head>`.

Source: https://v3.tailwindcss.com/docs/installation
