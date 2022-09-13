[![golangci-lint-status](https://github.com/mrjosh/helm-lint-ls/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/mrjosh/helm-lint-ls/actions/workflows/golangci-lint.yml)
[![Release](https://github.com/mrjosh/helm-lint-ls/actions/workflows/go-artifacts.yml/badge.svg)](https://github.com/mrjosh/helm-lint-ls/releases/latest)
![License](https://img.shields.io/github/license/mrjosh/helm-lint-ls)

<pre align="center">
      / / / /__  / /___ ___     / /   (_)___  / /_   / /   _____
     / /_/ / _ \/ / __  __ \   / /   / / __ \/ __/  / /   / ___/
    / __  /  __/ / / / / / /  / /___/ / / / / /_   / /___(__  ) 
 /_/ /_/\___/_/_/ /_/ /_/  /_____/_/_/ /_/\__/  /_____/____/
</pre>

## ⚠️ This project is still in early development. ⚠️

## Helm Lint Language Server Protocol
helm-lint-ls is [helm](https://github.com/helm/helm) lint language server protocol [LSP](https://microsoft.github.io/language-server-protocol/).

## Getting Started
### Vim Helm Plugin
You'll need [vim-helm](https://github.com/towolf/vim-helm) plugin installed before using helm_lint_ls, Try to install it vim:
```lua
Plug 'towolf/vim-helm'
```

### Download
* Download the latest helm_lint_ls executable file from [here](https://github.com/mrjosh/helm-lint-ls/releases/latest) and move it to your binaries directory 

* You can download it with curl, replace the {os} and {arch} variables
```bash
curl -L https://github.com/mrjosh/helm-lint-ls/releases/download/master/helm_lint_ls_{os}_{arch} --output /usr/local/bin/helm_lint_ls
```

### Make it executable
```bash
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

[![asciicast](https://asciinema.org/a/485522.svg)](https://asciinema.org/a/485522)

## Contributing
Thank you for considering contributing to HelmLintLs project!

## License
The HelmLintLs is open-source software licensed under the MIT license.
