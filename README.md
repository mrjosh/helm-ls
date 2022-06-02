<pre align="center">
      / / / /__  / /___ ___     / /   (_)___  / /_   / /   _____
     / /_/ / _ \/ / __  __ \   / /   / / __ \/ __/  / /   / ___/
    / __  /  __/ / / / / / /  / /___/ / / / / /_   / /___(__  ) 
 /_/ /_/\___/_/_/ /_/ /_/  /_____/_/_/ /_/\__/  /_____/____/
</pre>
### This project is under construction.
[![asciicast](https://asciinema.org/a/485522.svg)](https://asciinema.org/a/485522)

## About HelmLintLs
helm-lint-ls is [helm](https://github.com/helm/helm) lint language server protocol [LSP](https://microsoft.github.io/language-server-protocol/).

## Installation
Download the latest helm_lint_ls executable file from [here](https://github.com/mrjosh/helm-lint-ls/releases/latest) and move it to your binaries directory 

## Download it with curl
replace the {os} and {arch} variables in the url
```console
curl -L https://github.com/mrjosh/helm-lint-ls/releases/download/master/helm_lint_ls_{os}_{arch} --output /usr/local/bin/helm_lint_ls
```
### Make it executable
```console
chmod +x /usr/local/bin/helm_lint_ls
```

## nvim-lspconfig setup
```lua
local configs = require('lspconfig.configs')
local lspconfig = require('lspconfig')
local util = require('lspconfig.util')

if not configs.helm_lint_ls then
  configs.helm_lint_ls = {
    default_config = {
      cmd = {"helm_lint_ls", "serve"},
      filetypes = {'helm'},
      root_dir = function(fname)
        return util.root_pattern('Chart.yaml')(fname)
      end,
    },
  }
end

lspconfig.helm_lint_ls.setup {
  filetypes = {"helm"},
  cmd = {"helm_lint_ls", "serve"},
}
```

## Contributing
Thank you for considering contributing to HelmLintLs project!

## License
The HelmLintLs is open-source software licensed under the MIT license.

