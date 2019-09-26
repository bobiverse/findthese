```
Flags:
     --version  Displays the program version string.
  -h --help  Displays help with available flag, subcommand, and positional value parameters.
  -s --src  Source path of directory -- REQUIRED
  -u --url  URL endpoint to hit -- REQUIRED
  -m --method  HTTP Method to use (default: HEAD)
  -o --output  Output report to file (default: ./findthese.report)
     --depth  How deep go in folders. '0' no limit  (default: 0)
  -z --delay  Delay every request for N milliseconds (default: 250)
     --skip  Skip files with these extensions (default: [jquery css img images i18n po])
     --skip-ext  Skip files with these extensions (default: [.png .jpeg jpg Gif .CSS .less .sass])
     --skip-code  Skip responses with this response HTTP code (default: [404])
     --skip-size  Skip responses with this body size (default: [])
     --dir-only  Scan directories only

```


### TODO
- Mark placeholder for file to put in URL: https://example.com?f=^FILE^&auth=john
- [--mode=info|download] (default: info)
- file download path where to download all files
- On first success pause show content and press Y to continue
- read robots and .htaccess for paths
- Generate wordlist based on source directory
- Use `tor` by default (golang sockets transport)
-
