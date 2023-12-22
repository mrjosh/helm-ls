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
Helm-ls is a [helm](https://github.com/helm/helm) language server protocol [LSP](https://microsoft.github.io/language-server-protocol/) implementation.


<!-- vim-markdown-toc GFM -->

* [Demo](#demo)
* [Getting Started](#getting-started)
  * [Download](#download)
  * [Make it executable](#make-it-executable)
  * [Integration with yaml-language-server](#integration-with-yaml-language-server)
* [Configuration options](#configuration-options)
  * [LSP Server](#lsp-server)
  * [yaml-language-server config](#yaml-language-server-config)
  * [Emacs eglot setup](#emacs-eglot-setup)
* [Contributing](#contributing)
* [License](#license)

<!-- vim-markdown-toc -->

## Demo
[![asciicast](https://asciinema.org/a/485522.svg)](https://asciinema.org/a/485522)

## Getting Started

### Download
* Download the latest helm_ls executable file from [here](https://github.com/mrjosh/helm-ls/releases/latest) and move it to your binaries directory 

* You can download it with curl, replace the {os} and {arch} variables
```bash
curl -L https://github.com/mrjosh/helm-ls/releases/download/master/helm_ls_{os}_{arch} --output /usr/local/bin/helm_ls
```

If you are using neovim with [mason](https://github.com/williamboman/mason.nvim) you can also install it with mason.

### Make it executable
```bash
chmod +x /usr/local/bin/helm_ls
```

### Integration with [yaml-language-server](https://github.com/redhat-developer/yaml-language-server)
Helm-ls will use yaml-language-server to provide additional capabilities, if it is installed.
This feature is expermiental, you can disable it in the config ([see](#configuration-options)).
Having a broken template syntax (e.g. while your are stil typing) will cause diagnostics from yaml-language-server to be shown as errors.

To install it using npm run (or use your preferred package manager):
```bash
npm install --global yaml-language-server
```

The default kubernetes schema of yaml-language-server will be used for all files. You can overwrite which schema to use in the config ([see](#configuration-options)).
If you are for example using CRDs that are not included in the default schema, you can overwrite the schema using a comment
to use the schemas from the [CRDs-catalog](https://github.com/datreeio/CRDs-catalog).

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/datreeio/CRDs-catalog/main/keda.sh/scaledobject_v1alpha1.json
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
...
```

## Configuration options

You can configure helm-ls with lsp workspace configurations.

### LSP Server

- **Log Level**: Adjust log verbosity.

### yaml-language-server config

  - **Enable yaml-language-server**: Toggle support of this feature.
  - **Path to yaml-language-server**: Specify the executable location.
  - **Diagnostics Settings**:
    - **Limit**: Number of displayed diagnostics per file.
    - **Show Directly**: Show diagnostics while typing.

  - **Additional Settings** (see [yaml-language-server](https://github.com/redhat-developer/yaml-language-server#language-server-settings)):
    - **Schemas**: Define YAML schemas.
    - **Completion**: Enable code completion.
    - **Hover Information**: Enable hover details.

  ### Default Configuration

  ```lua
  settings = {
    ['helm-ls'] = {
      logLevel = "debug",
      yamlls = {
        enabled = true,
        diagnosticsLimit = 50,
        showDiagnosticsDirectly = false,
        path = "yaml-language-server",
        config = {
          schemas = {
            kubernetes = "**",
          },
          completion = true,
          hover = true,
          -- any other config: https://github.com/redhat-developer/yaml-language-server#language-server-settings
        }
      }
    }
  }
  ```

  ## Editor Config examples

  ### Neovim (using nvim-lspconfig)
  #### Vim Helm Plugin
  You'll need [vim-helm](https://github.com/towolf/vim-helm) plugin installed before using helm_ls, to install it using vim-plug (or use your preferred plugin manager):
  ```lua
  Plug 'towolf/vim-helm'
  ```

  #### Setup laguage server
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
See [examples/nvim/init.lua](https://github.com/mrjosh/helm-ls/blob/master/examples/nvim/init.lua) for an
complete example, which also includes yaml-language-server.


###  Emacs eglot setup

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
