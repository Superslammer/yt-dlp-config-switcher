# yt-dlp-config-switcher
A wrapper around yt-dlp that adds the ability to use different config files

# Usage
`yt-dlp-switcher -c <config-name> <youtube url>`

`<config-name>` is the name of the config file in the `yt-dlp configs` folder.

`<youtube url>` is the url to the youtube video you are trying to download

If `-c <config-name>` is omitted the default config file is used

## Example
To use it with a specific config file fx audio.conf

`yt-dlp-switcher -c audio https://www.youtube.com/watch?v=UCekG0O21wQ`

# Build
In the root directory run `go build`
