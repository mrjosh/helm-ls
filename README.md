[![Lint](https://github.com/mrjosh/helm-ls/actions/workflows/lint.yml/badge.svg)](https://github.com/mrjosh/helm-ls/actions/workflows/lint.yml)
[![Tests](https://github.com/mrjosh/helm-ls/actions/workflows/tests.yml/badge.svg)](https://github.com/mrjosh/helm-ls/actions/workflows/tests.yml)
[![Release](https://github.com/mrjosh/helm-ls/actions/workflows/artifacts.yml/badge.svg)](https://github.com/mrjosh/helm-ls/releases/latest)
![License](https://img.shields.io/github/license/mrjosh/helm-ls)

<pre align="center">
  /\  /\___| |_ __ ___   / / ___ 
 / /_/ / _ \ | '_ ` _ \ / / / __|
/ __  /  __/ | | | | | / /__\__ \
\/ /_/ \___|_|_| |_| |_\____/___/
</pre>

## Helm Language Server Protocol
helm-ls is [helm](https://github.com/helm/helm) language server protocol [LSP](https://microsoft.github.io/language-server-protocol/).

## Getting Started
### Vim Helm Plugin
You'll need [vim-helm](https://github.com/towolf/vim-helm) plugin installed before using helm_ls, Try to install it vim:
```lua
Plug 'towolf/vim-helm'
```

### Download
* Download the latest helm_ls executable file from [here](https://github.com/mrjosh/helm-ls/releases/latest) and move it to your binaries directory 

* You can download it with curl, replace the {os} and {arch} variables
```bash
curl -L https://github.com/mrjosh/helm-ls/releases/download/master/helm_ls_{os}_{arch} --output /usr/local/bin/helm_ls
```

### Make it executable
```bash
chmod +x /usr/local/bin/helm_ls
```

## nvim-lspconfig setup
```lua
local lspconfig = require('lspconfig')

lspconfig.helm_ls.setup {
  settings = {
    ['helm-ls'] = {
      yamlls = {
        path = "yaml-language-server",
      }
    }
  }
}
```

[![asciicast](https://asciinema.org/a/485522.svg)](https://asciinema.org/a/485522)

## Emacs eglot setup

Integrating helm-ls with [eglot](https://github.com/joaotavora/eglot) for emacs consists of two steps: wiring up Helm template files into a specific major mode and then associating that major mode with `helm_ls` via the `eglot-server-programs` variable.
The first step is necessary because without a Helm-specific major mode, using an existing major mode like `yaml-mode` for `helm_ls` in `eglot-server-programs` may invoke the language server for other, non-Helm yaml files.

For example, the following elisp snippet demonstrates how to use this language server after installing it as explained in [Getting Started](#getting-started).
Assuming that you leverage `use-package` for package management:

```elisp
;; ...ensure that your package manager of choice is setup before
;; installing packages, and then

;; Install yaml-mode
(use-package yaml-mode)

;; Create a derived major-mode based on yaml-mode
(define-derived-mode helm-mode yaml-mode "helm"
  "Major mode for editing kubernetes helm templates")

(use-package eglot
  ; Any other existing eglot configuration plus the following:
  :hook
  ; Run eglot in helm-mode buffers
  (helm-mode . eglot-ensure)
  :config
  ; Run `helm_ls serve` for helm-mode buffers
  (add-to-list 'eglot-server-programs '(helm-mode "helm_ls" "serve")))
```

Invoke `M-x helm-mode` in a Helm template file to begin using helm-ls as a backend for eglot.
Alternatively, you can include a comment such as the following at the top of Helm yaml files to automatically enter `helm-mode`:

    # -*- mode: helm -*-

## Contributing
Thank you for considering contributing to HelmLs project!

## License
The HelmLs is open-source software licensed under the MIT license.
