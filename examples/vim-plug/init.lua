-- a minimal example config for setting up neovim with helm-ls and yamlls using the plugin manager vim-plug
-- does not cover syntax highlighting, see ../nvim/init.lua for that
-- test it with: nvim -u init.lua
-- make sure to follow the vim-plug installation instructions: https://github.com/junegunn/vim-plug

local Plug = vim.fn["plug#"]

-- install required plugins
vim.call("plug#begin")

Plug("https://github.com/neovim/nvim-lspconfig")
Plug("https://github.com/qvalentin/helm-ls.nvim")

vim.call("plug#end")

-- setup helm-ls.nvim
require("helm-ls").setup({
	conceal_templates = {
		-- enable the replacement of templates with virtual text of their current values
		enabled = false, -- tree-sitter must be setup for this feature
	},
	indent_hints = {
		-- enable hints for indent and nindent functions
		enabled = false, -- tree-sitter must be setup for this feature
	},
})

local lspconfig = vim.lsp.config
vim.lsp.enable("helm_ls")
-- setup helm-ls
lspconfig("helm_ls", {
	settings = {
		["helm-ls"] = {
			yamlls = {
				path = "yaml-language-server",
			},
		},
	},
})

-- enable yamlls
vim.lsp.enable("yamlls")
-- optional: configure yamlls options
lspconfig("yamlls", {})
