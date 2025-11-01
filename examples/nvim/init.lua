-- a minimal example config for setting up neovim with helm-ls and yamlls
-- uses tree-sitter for syntax highlighting
-- test it with: nvim -u init.lua

-- setup lazy plugin manager
local lazypath = vim.fn.stdpath("data") .. "/lazy/lazy.nvim"
if not vim.loop.fs_stat(lazypath) then
	vim.fn.system({
		"git",
		"clone",
		"--filter=blob:none",
		"https://github.com/folke/lazy.nvim.git",
		"--branch=stable", -- latest stable release
		lazypath,
	})
end
vim.opt.rtp:prepend(lazypath)
vim.g.mapleader = " "
-- end of lazy setup

-- install required plugins
require("lazy").setup({
	-- use tree-sitter for syntax highlighting
	{
		"nvim-treesitter/nvim-treesitter",
		branch = "main",
		lazy = false,
		build = ":TSUpdate",
	},
	-- alternativly you can use towolf/vim-helm for basic syntax highlighting
	-- { "towolf/vim-helm",       ft = "helm" }, -- ft = 'helm' is important to not start yamlls
	{
		"qvalentin/helm-ls.nvim",
		ft = "helm",
		opts = {
			conceal_templates = {
				-- enable the replacement of templates with virtual text of their current values
				enabled = true, -- tree-sitter must be setup for this feature
			},
			indent_hints = {
				-- enable hints for indent and nindent functions
				enabled = true, -- tree-sitter must be setup for this feature
			},
		},
	},
	{ "neovim/nvim-lspconfig", event = { "BufReadPre", "BufNewFile", "BufEnter" } },
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

-- setup treesitter for syntax highlighting
require("nvim-treesitter").setup({
	-- A list of parser names, or "all" (the listed parsers MUST always be installed)
	ensure_installed = { "yaml", "helm" },
	-- Install parsers synchronously (only applied to `ensure_installed`)
	sync_install = true,
	-- Automatically install missing parsers when entering buffer
	-- Recommendation: set to false if you don't have `tree-sitter` CLI installed locally
	auto_install = true,
	highlight = {
		enable = true,
	},
})
