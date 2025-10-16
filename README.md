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

## Helm Language Server

Helm-ls is a [helm](https://github.com/helm/helm) language server protocol [LSP](https://microsoft.github.io/language-server-protocol/) implementation.

<!-- vim-markdown-toc GFM -->

- [Demo](#demo)
- [Getting Started](#getting-started)
  - [Installation with a package manager](#installation-with-a-package-manager)
    - [Homebrew](#homebrew)
    - [Nix](#nix)
    - [Arch Linux](#arch-linux)
    - [Windows](#windows)
    - [mason (neovim)](#mason-neovim)
  - [Manual download](#manual-download)
    - [Make it executable](#make-it-executable)
  - [Integration with yaml-language-server](#integration-with-yaml-language-server)
    - [Template files](#template-files)
    - [Values files](#values-files)
    - [Install](#install)
    - [Custom Schemas](#custom-schemas)
  - [Dependency Charts](#dependency-charts)
- [Configuration options](#configuration-options)
  - [General](#general)
  - [Values Files](#values-files-1)
  - [Helm Lint](#helm-lint)
  - [yaml-language-server config](#yaml-language-server-config)
  - [Default Configuration](#default-configuration)
- [Editor Config examples](#editor-config-examples)
  - [Neovim](#neovim)
    - [Filetype detection](#filetype-detection)
    - [nvim-lspconfig setup](#nvim-lspconfig-setup)
    - [coc.nvim setup](#cocnvim-setup)
  - [VSCode](#vscode)
  - [Zed](#zed)
  - [Emacs eglot setup](#emacs-eglot-setup)
- [Features and Demos](#features-and-demos)
- [Contributing](#contributing)
- [License](#license)

<!-- vim-markdown-toc -->

## Demo

[![asciicast](https://asciinema.org/a/485522.svg)](https://asciinema.org/a/485522)

## Getting Started

### Installation with a package manager

Helm-ls is currently available as a package for some package managers.

[![Packaging status](https://repology.org/badge/vertical-allrepos/helm-ls.svg)](https://repology.org/project/helm-ls/versions)

These are some of the supported package managers. Thanks to everyone who packaged it!

#### Homebrew

If you are using MacOS or Linux with [Homebrew](https://brew.sh/) you can install it with brew.

```bash
brew install helm-ls
```

#### Nix

```bash
nix-shell -p helm-ls
```

#### Arch Linux

You can install it from the [aur](https://aur.archlinux.org/packages/helm-ls/) using your preferred aur helper, e.g. yay:

```bash
yay -S helm-ls
# or
yay -S helm-ls-bin
```

#### Windows

You can use [scoop](https://scoop.sh/) to install it:

```powershell
scoop bucket add extras
scoop install extras/helm-ls
```

#### mason (neovim)

If you are using neovim with [mason](https://github.com/williamboman/mason.nvim) you can also install it with mason.

```vim
:MasonInstall helm-ls
```

### Manual download

- Download the latest helm_ls executable file from [here](https://github.com/mrjosh/helm-ls/releases/latest) and move it to your binaries directory

- You can download it with curl, replace the {os} and {arch} variables

```bash
curl -L https://github.com/mrjosh/helm-ls/releases/download/master/helm_ls_{os}_{arch} --output /usr/local/bin/helm_ls
```

#### Make it executable

```bash
chmod +x /usr/local/bin/helm_ls
```

### Integration with [yaml-language-server](https://github.com/redhat-developer/yaml-language-server)

Helm-ls will use yaml-language-server to provide additional capabilities, if it is installed.

#### Template files

Helm-ls will convert the gotemplate files in the templates directory to yaml and process them with yaml-language-server.

> [!WARNING]
>
> This feature is experimental, you can disable it in the config ([see](#configuration-options)) if you are getting a lot of errors beginning with `Yamlls:`.
> Having a broken template syntax (e.g. while your are still typing) will also cause diagnostics from yaml-language-server to be shown as errors.

#### Values files

Helm-ls will generate json-schemas for all values.\*yaml files and use yaml-language-server to provide autocompletion.
This feature is currently in beta, see [this discussion](https://github.com/mrjosh/helm-ls/issues/61#issuecomment-2927585818) for details.

#### Install

```bash
npm install --global yaml-language-server
```

To install it using npm run (or use your preferred package manager):

The default kubernetes schema of yaml-language-server will be used for all files. You can overwrite which schema to use in the config ([see](#configuration-options)).
If you are for example using CRDs that are not included in the default schema, you can overwrite the schema using a comment
to use the schemas from the [CRDs-catalog](https://github.com/datreeio/CRDs-catalog).

#### Custom Schemas

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/datreeio/CRDs-catalog/main/keda.sh/scaledobject_v1alpha1.json
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
```

### Dependency Charts

Helm-ls can process dependency charts to provide autocompletion, hover etc. with values from the dependencies.
For this the dependency charts have to be downloaded. Run the following command in your project to download them:

```bash
helm dependency build
```

## Configuration options

You can configure helm-ls with lsp workspace configurations.

### General

- **Log Level**: Adjust log verbosity.

### Values Files

- **Main Values File**: Path to the main values file (values.yaml per default)
- **Lint Overlay Values File**: Path to the lint overlay values file, which will be merged with the main values file for linting
- **Additional Values Files Glob Pattern**: Pattern for additional values files, which will be shown for completion and hover

### Helm Lint

- **Enabled**: Allows to disable the diagnostics gathered with helm lint.
- **ignoredMessages**: A list of lint messages to ignore. You can for example add "icon is recommended" here.

### yaml-language-server config

- **Enable yaml-language-server**: Toggle support of this feature.
- **EnabledForFilesGlob**: A glob pattern defining for which files yaml-language-server should be enabled.
- **Path to yaml-language-server**: Specify the executable location. Can be a string or an array. Can also be set with the environment variable `YAMLLS_PATH` using a comma separated list or a JSON array.
- **initTimeoutSeconds**: The timeout in seconds for the initialization of yamlls. (Increase if you get an error log like "Error initializing yamlls context deadline exceeded")
- **Diagnostics Settings**:

  - **Limit**: Number of displayed diagnostics per file. Set this to 0 to disable all diagnostics from yaml-language-server but keep other features such as hover.
  - **Show Directly**: Show diagnostics while typing.

- **Additional Settings** (see [yaml-language-server](https://github.com/redhat-developer/yaml-language-server#language-server-settings)):
  - **Schemas**: Define YAML schemas.
  - **Completion**: Enable code completion.
  - **Hover Information**: Enable hover details.

### Default Configuration

```lua
settings = {
  ['helm-ls'] = {
    logLevel = "info",
    valuesFiles = {
      mainValuesFile = "values.yaml",
      lintOverlayValuesFile = "values.lint.yaml",
      additionalValuesFilesGlobPattern = "values*.yaml"
    },
    helmLint = {
      enabled = true,
      ignoredMessages = {},
    },
    yamlls = {
      enabled = true,
      enabledForFilesGlob = "*.{yaml,yml}",
      diagnosticsLimit = 50,
      showDiagnosticsDirectly = false,
      path = "yaml-language-server", -- or something like { "node", "yaml-language-server.js" }
      initTimeoutSeconds = 3,
      config = {
        schemas = {
          kubernetes = "templates/**",
        },
        completion = true,
        hover = true,
        -- any other config from https://github.com/redhat-developer/yaml-language-server#language-server-settings
      }
    }
  }
}
```

## Editor Config examples

### Neovim

#### Filetype detection

To get filetype detection working, you can use one of the folowing plugins:

- [helm-ls.nvim](https://github.com/qvalentin/helm-ls.nvim): **recommended**, requires [tree-sitter](https://github.com/ngalaiko/tree-sitter-go-template?tab=readme-ov-file#neovim-integration-using-nvim-treesitter) for syntax highlighting. Also provides some additional features.
- [vim-helm](https://github.com/towolf/vim-helm): known to cause problems with yaml-language-server when used with another plugin manger than lazy

install it using lazy (or use your preferred plugin manager):

```lua
{ "qvalentin/helm-ls.nvim", ft = "helm" }
-- or { "towolf/vim-helm", ft = "helm" },
-- or even both if you do not want to use tree-sitter for syntax highlighting
```

#### nvim-lspconfig setup

Add the following to your neovim lua config:

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
complete example using lazy, or [examples/vim-plug/init.lua](https://github.com/mrjosh/helm-ls/blob/master/examples/vim-plug/init.lua) for vim-plug.
The examples also include the setup for yaml-language-server.

> [!TIP]
>
> If you are using [AstroNvim](https://github.com/AstroNvim/AstroNvim) you can just install the [astrocommunity](https://github.com/AstroNvim/astrocommunity) helm pack
> or if using [LazyVim](https://github.com/LazyVim/LazyVim) its [LazyVimHelm](https://github.com/LazyVim/LazyVim) plugin.

#### coc.nvim setup

You can also use [coc.nvim](https://github.com/neoclide/coc.nvim) to set up the language server.
You will need to configure the use of `helm_ls` in the `langageserver` section of your `coc-settings.json` file.

Open Neovim and type the command `:CocConfig` to access the configuration file. Find the `langageserver` section and add this configuration:

```json
"languageserver": {
  "helm": {
    "command": "helm_ls",
    "args": ["serve"],
    "filetypes": ["helm", "helmfile"],
    "rootPatterns": ["Chart.yaml"]
  }
}
```

Save the configuration file and then either restart Neovim or type `:CocRestart` to restart the language server.

### VSCode

Check out the [helm-ls-vscode extension](https://github.com/qvalentin/helm-ls-vscode) for more details.

### Zed

Setup filetypes as described in the [Zed Docs](https://zed.dev/docs/languages/helm) and install the [helm.zed extension](https://github.com/cabrinha/helm.zed).

### Emacs eglot setup

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

## Features and Demos

<details>
  <summary>
	<b>Hover</b>
  </summary>

<video alt="demo for hover" src="https://github.com/user-attachments/assets/48413b5b-aedf-4735-aeca-aff32553f3fd"></video>

| Language Construct (or filetype)      | Example Effect                                                                                                                                   |
| ------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| Values                                | `.Values.replicaCount` shows the value of `replicaCount` in the values.yaml files.                                                               |
| Built-In-Objects                      | `.Chart.Name` shows the name of the Chart.                                                                                                       |
| Includes                              | `include "example.labels"` shows the definition of the template.                                                                                 |
| Functions                             | `add` shows the docs of the add function.                                                                                                        |
| Yaml in Templates                     | `Kind` shows the docs from the yaml-schema (via yaml-language-server).                                                                           |
| [values.\*.yaml files](#values-files) | Docs from the generated schema (via yaml-language-server), YAML path, values from other values files either from the same Chart or other Charts. |

</details>

<details>
  <summary>
	<b>Autocomplete</b>
  </summary>

<video alt="Demo for autocompletion" src="https://github.com/user-attachments/assets/15c57a0a-4a17-48b4-9861-a324bcfa2158"></video>

| Language Construct (or filetype)      | Effect                                                                     |
| ------------------------------------- | -------------------------------------------------------------------------- |
| Values                                | Values from `values*.yaml` files (including child/parent Charts).          |
| Built-In-Objects                      | Values from `Chart`, `Release`, `Files`, `Capabilities`, `Template`.       |
| Includes                              | Available includes (including child/parent Charts).                        |
| Functions                             | Functions from gotemplate and helm.                                        |
| Yaml in Templates                     | Values from the yaml-schema (via yaml-language-server).                    |
| [values.\*.yaml files](#values-files) | Values from other values files either from the same Chart or other Charts. |

</details>

<details>
  <summary>
	<b>Go-To-Definition/References</b>
  </summary>

<video alt="Demo for definition and references" src="https://github.com/user-attachments/assets/e49769e9-4ddb-4b05-b075-645a9f9b9937"></video>

| Language Construct                    | Effect                                                                                                 |
| ------------------------------------- | ------------------------------------------------------------------------------------------------------ |
| Values                                | Go to `values*.yaml` files for template references (including child/parent Charts) or other templates. |
| Built-In-Objects                      | Go to `Chart.yaml` for `Chart.*`.                                                                      |
| Includes                              | Go to definition/references of templates (including child/parent Charts).                              |
| [values.\*.yaml files](#values-files) | Go to other `values*.yaml` files (definitions) or templates using the values (references).             |

</details>

<details>
  <summary>
	<b>Symbol</b>
  </summary>

Can show a breadcrumb of the yaml path of the current position (via yaml-language-server).
![Demo for Symbol](https://github.com/user-attachments/assets/0b8a9fc4-4625-4641-a296-8aedb48496e9)

</details>

<details>
  <summary>
	<b>Linting</b>
  </summary>

Diagnostics from both helm lint and yaml-language-server.
![Demo of Linting](https://github.com/user-attachments/assets/58e90dd4-2fe5-40f5-a9a7-adec6c890a0c)

</details>

## Contributing

Thank you for considering contributing to Helm-ls project!

## License

The Helm-ls is open-source software licensed under the MIT license.

Part of the documentation that is included in helm-ls is copied from the Go standard library. The original license is included in the files containing the documentation.
