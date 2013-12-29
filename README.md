# aria2fwd

aria2fwd is a helper application to send URLs to a remote (or local) aria2 instance. It is designed with a view to being called directly from a web browser, but may also be used from the command line.

## Usage

From the command line: `aria2fwd URI` where URI is an HTTP URL, Magnet URI or the filepath of a .torrent file.

aria2fwd is designed to be used from a web browser. Set it as the default handler for Magnet URIs and downloaded torrent files to have them sent to the configured aria2 instance. 

## Configuration

Edit config.json or run `aria2fwd -c` to generate a config file. 

aria2fwd can be configured to run a command when a download is successfully registered or when an error occurs. The Linux-centric example below opens a web browser when a download is registered, or displays a notification-daemon message in the event of an error.

```json
{
/* address and port of the aria2 server */
"Addr":"127.0.0.1:6800",
/* command to run on success; %s is replaced with the download ID */
"Success":"xdg-open http://my-fancy-aria2.web.ui/id/%s",
/* command to run on failure; %s is replaced with an error message */
"Error":"notify-send 'aria2 error' '%s'"
}
```

